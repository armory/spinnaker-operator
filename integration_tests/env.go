package integration_tests

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	KubeconfigVar                        = "KUBECONFIG"
	OperatorImageVar                     = "OPERATOR_IMAGE"
	HalyardImageVar                      = "HALYARD_IMAGE"
	BucketVar                            = "S3_BUCKET"
	BucketRegionVar                      = "AWS_DEFAULT_REGION"
	BucketAccessKeyIdVar                 = "AWS_ACCESS_KEY_ID"
	BucketSecretAccessKeyVar             = "AWS_SECRET_ACCESS_KEY"
	BucketSessionTokenVar                = "AWS_SESSION_TOKEN"
	MaxChecksWaitingForAccountsAvailable = 20 // 20 * 2s = 40 seconds
)

type Defaults struct {
	HalyardImageDefault            string
	BucketDefault                  string
	BucketRegionDefault            string
	BucketAccessKeyIdDefault       string
	BucketSecretAccessKeyIdDefault string
	OperatorKustomizeBase          string
	CRDManifests                   string
	OperatorImageDefault           string
}

var envLock sync.Mutex
var baseEnv = TestEnv{}
var opClusterLock sync.Mutex
var opCluster = Operator{}

// TestEnv holds information about the kubernetes cluster used for tests
type TestEnv struct {
	Operator    Operator
	SpinDeckUrl string
	SpinGateUrl string
	Vars        Vars
}

// Operator holds information about the operator installation
type Operator struct {
	KustomizationPath string
	Namespace         string
	PodName           string
}

type Account struct {
	Name  string   `json:"name,omitempty"`
	Type  string   `json:"type,omitempty"`
	Types []string `json:"types,omitempty"`
}

// Vars are variables used in kustomize templates
type Vars struct {
	Kubeconfig        string
	OperatorImage     string
	HalyardImage      string
	S3Bucket          string
	S3BucketRegion    string
	S3AccessKeyId     string
	S3SecretAccessKey string
	S3SessionToken    string
	SpinNamespace     string
}

// CommonSetup creates a new environment context, initializing common settings for all tests
func CommonSetup(d Defaults, t *testing.T) *TestEnv {
	envLock.Lock()
	defer envLock.Unlock()
	if baseEnv.Vars.Kubeconfig != "" {
		t.Logf("Environment already initialized")
	} else {
		baseEnv = TestEnv{
			Vars: resolveEnvVars(d, t),
		}
		SubstituteOverlayVars(d.OperatorKustomizeBase, baseEnv.Vars, t)
		if t.Failed() {
			return nil
		}
		baseEnv.InstallCrds(d, t)
		SubstituteOverlayVars("testdata/spinnaker/base", baseEnv.Vars, t)
	}
	return &TestEnv{
		Vars: baseEnv.Vars,
	}
}

func resolveEnvVars(d Defaults, t *testing.T) Vars {
	k := os.Getenv(KubeconfigVar)
	if k == "" {
		t.Logf("%s env var not set, using default", KubeconfigVar)
		home, err := os.UserHomeDir()
		if !assert.Nil(t, err, "error getting user home") {
			return Vars{}
		}
		k = fmt.Sprintf("%s/.kube/config", home)
	}
	t.Logf("Using kubeconfig %s", k)

	op := os.Getenv(OperatorImageVar)
	if op == "" {
		t.Logf("%s env var not set, using default", OperatorImageVar)
		op = d.OperatorImageDefault
	}
	t.Logf("Using operator image %s", op)

	h := os.Getenv(HalyardImageVar)
	if h == "" {
		t.Logf("%s env var not set, using default", HalyardImageVar)
		h = d.HalyardImageDefault
	}
	t.Logf("Using halyard image %s", h)

	b := os.Getenv(BucketVar)
	if b == "" {
		t.Logf("%s env var not set, using default", BucketVar)
		b = d.BucketDefault
	}
	t.Logf("Using bucket %s", b)

	r := os.Getenv(BucketRegionVar)
	if r == "" {
		t.Logf("%s env var not set, using default", d.BucketRegionDefault)
		r = d.BucketRegionDefault
	}
	t.Logf("Using bucekt region %s", r)

	a := os.Getenv(BucketAccessKeyIdVar)
	if a == "" {
		t.Fatalf("%s env var not set", BucketAccessKeyIdVar)
	}
	s := os.Getenv(BucketSecretAccessKeyVar)
	if s == "" {
		t.Fatalf("%s env var not set", BucketSecretAccessKeyVar)
	}
	return Vars{
		Kubeconfig:        k,
		OperatorImage:     op,
		HalyardImage:      h,
		S3Bucket:          b,
		S3BucketRegion:    r,
		S3AccessKeyId:     a,
		S3SecretAccessKey: s,
		S3SessionToken:    strings.TrimSpace(os.Getenv(BucketSessionTokenVar)),
	}
}

func (e *TestEnv) KubectlPrefix() string {
	return fmt.Sprintf("kubectl --kubeconfig=%s", e.Vars.Kubeconfig)
}

func (e *TestEnv) Cleanup(t *testing.T) {
	e.DeleteOperator(t)
}

func InstallCrdsAndOperator(spinNs string, isClusterMode bool, d Defaults, t *testing.T) (e *TestEnv) {
	e = CommonSetup(d, t)
	if t.Failed() {
		return
	}
	e.Vars.SpinNamespace = spinNs
	if isClusterMode {
		opClusterLock.Lock()
		defer opClusterLock.Unlock()
		if opCluster.KustomizationPath != "" {
			t.Logf("Operator in cluster mode already installed")
		} else {
			opCluster = e.InstallOperator(isClusterMode, t)
		}
		e.Operator = opCluster
	} else {
		e.Operator = e.InstallOperator(isClusterMode, t)
	}
	return
}

func (e *TestEnv) InstallCrds(d Defaults, t *testing.T) bool {
	ApplyManifest("default", d.CRDManifests, e, t)
	_, _ = RunCommand("rm -rf ~/.kube/http-cache/ && rm -rf ~/.kube/cache/", t)
	RunCommandAndAssert(fmt.Sprintf("%s get spinsvc", e.KubectlPrefix()), t)
	RunCommandAndAssert(fmt.Sprintf("%s get spinnakeraccounts", e.KubectlPrefix()), t)
	return !t.Failed()
}

func (e *TestEnv) InstallOperator(isCluster bool, t *testing.T) Operator {
	opKustPath := "testdata/operator/overlay_basicmode"
	if isCluster {
		opKustPath = "testdata/operator/overlay_clustermode"
	}
	op := Operator{
		KustomizationPath: opKustPath,
		Namespace:         RandomString("operator"),
	}
	LogMainStep(t, "Installing CRDs and operator in namespace %s", op.Namespace)
	SubstituteOverlayVars(opKustPath, op, t)
	if !CreateNamespace(op.Namespace, e, t) {
		return Operator{}
	}
	if !ApplyKustomizeAndAssert(op.Namespace, opKustPath, e, t) {
		return Operator{}
	}
	if !WaitForDeploymentToStabilize(op.Namespace, "spinnaker-operator", e, t) {
		return Operator{}
	}
	p := RunCommandAndAssert(fmt.Sprintf("%s -n %s get pods | grep spinnaker-operator | awk '{print $1}'", e.KubectlPrefix(), op.Namespace), t)
	op.PodName = strings.TrimSpace(p)
	LogMainStep(t, "CRDs and operator installed")
	return op
}

func (e *TestEnv) DeleteOperator(t *testing.T) {
	t.Logf("Deleting operator...")
	DeleteNamespace(e.Operator.Namespace, e, t)
}

func (e *TestEnv) InstallSpinnaker(ns, kustPath string, t *testing.T) bool {
	LogMainStep(t, "Installing spinnaker in namespace %s", ns)
	if !CreateNamespace(ns, e, t) {
		return !t.Failed()
	}
	e.SpinDeckUrl, e.SpinGateUrl = DeploySpinnaker(ns, kustPath, e, t)
	if t.Failed() {
		return !t.Failed()
	}
	LogMainStep(t, "Spinnaker installed successfully")
	return !t.Failed()
}

func (e *TestEnv) VerifyAccountsExist(endpoint string, t *testing.T, accts ...Account) bool {
	LogMainStep(t, "Verifying spinnaker accounts")
	o := ""
	var credentials []Account
	for counter := 0; counter < MaxChecksWaitingForAccountsAvailable; counter++ {
		o = ExecuteGetRequest(fmt.Sprintf("%s%s", e.SpinGateUrl, endpoint), t)
		if t.Failed() {
			return !t.Failed()
		}
		found := 0
		credentials = make([]Account, 0)
		if assert.Nil(t, json.Unmarshal([]byte(o), &credentials)) {
			for _, a := range accts {
				for _, c := range credentials {
					if a.Type != "" && a.Type == c.Type && a.Name == c.Name {
						found++
						break
					}
					if a.Types != nil && len(a.Types) > 0 && len(c.Types) > 0 && a.Types[0] == c.Types[0] && a.Name == c.Name {
						found++
						break
					}
				}
			}
		}
		if len(accts) == found {
			return !t.Failed()
		}
		time.Sleep(2 * time.Second)
	}
	t.Errorf("Unable to find all accounts in spinnaker. Expected: %v but found: %v. Requested url: %s, response: %s", accts, credentials, endpoint, o)
	return !t.Failed()
}

func (e *TestEnv) GenerateSpinFiles(kustPath, name, filePath string, t *testing.T) bool {
	f := `
# This file is automatically generated by integration tests (env.go), any changes will be lost
apiVersion: spinnaker.io/v1alpha2
kind: SpinnakerService
metadata:
  name: spinnaker
spec:
  spinnakerConfig:
    files:
      %s: |
%s
`
	// read and indent file
	h, err := os.Open(filePath)
	if !assert.Nil(t, err) {
		return !t.Failed()
	}
	s := bufio.NewScanner(h)
	indentedFile := ""
	for s.Scan() {
		indentedFile += fmt.Sprintf("        %s\n", s.Text())
	}
	if !assert.Nil(t, s.Err()) {
		return !t.Failed()
	}

	f = fmt.Sprintf(f, name, string(indentedFile))
	err = ioutil.WriteFile(filepath.Join(kustPath, "files.yml"), []byte(f), os.ModePerm)
	assert.Nil(t, err, "unable to generate files.yml file")

	UpdateControlPlaneHost(filepath.Join(kustPath, "files.yml"), t)

	return !t.Failed()
}
