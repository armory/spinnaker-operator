package integration_tests

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
)

const (
	NsOperatorCluster = "test-operator-cluster-mode" // This value needs to be in sync with operator RoleBinding manifest
	NsOperatorBasic   = "test-spinnaker-basic-mode"
	NsSpinnaker1      = "test-spinnaker-cluster-mode"

	KubeconfigVar        = "KUBECONFIG"
	OperatorImageVar     = "OPERATOR_IMAGE"
	OperatorImageDefault = "armory/spinnaker-operator:dev"
	HalyardImageVar      = "HALYARD_IMAGE"
	HalyardImageDefault  = "armory/halyard:operator-0.3.x"
	CleanStartVar        = "CLEAN_START"

	OperatorKustomizeBase   = "testdata/operator/base"
	OperatorSourceManifests = "../deploy/operator/cluster"
	CRDManifests            = "../deploy/crds"
)

var envLock sync.Mutex
var envInitialized = false
var operatorRunsByNamespace = map[string]bool{}

// TestEnv holds information about the kubernetes cluster used for tests
type TestEnv struct {
	KubeconfigPath string
	Operator       Operator
	SpinDeckUrl    string
	SpinGateUrl    string
}

// Operator holds information about the operator installation
type Operator struct {
	KustomizationPath string
	Namespace         string
	OperatorImage     string
	HalyardImage      string
}

type Account struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// NewEnv creates a new environment context, with the given operator namespace and path pointing to a kustomize folder
// with operator manifests
func NewEnv(opNs, opKust string, t *testing.T) *TestEnv {
	envLock.Lock()
	defer envLock.Unlock()
	k := os.Getenv(KubeconfigVar)
	if k == "" {
		t.Logf("%s env var not set, using default", KubeconfigVar)
		home, err := os.UserHomeDir()
		if !assert.Nil(t, err, "error getting user home") {
			return nil
		}
		k = fmt.Sprintf("%s/.kube/config", home)
	}
	t.Logf("Using kubeconfig %s", k)
	e := &TestEnv{
		KubeconfigPath: k,
		Operator: Operator{
			Namespace:         opNs,
			KustomizationPath: opKust,
		},
	}
	if envInitialized {
		t.Logf("Environment already initialized")
		return e
	}
	envInitialized = true
	generateKustomizeBase(t)
	if t.Failed() {
		return nil
	}
	cleanEnvIfNeeded(e, t)
	return e
}

func cleanEnvIfNeeded(e *TestEnv, t *testing.T) {
	c := os.Getenv(CleanStartVar)
	if c == "" {
		return
	}
	b, err := strconv.ParseBool(c)
	if err != nil {
		t.Logf("Unable to parse a bool from env var %s: %s, ignoring and not cleaning environment", CleanStartVar, c)
		return
	}
	if b {
		DeleteNamespace(NsOperatorCluster, e, t)
		DeleteNamespace(NsOperatorBasic, e, t)
		DeleteNamespace(NsSpinnaker1, e, t)
	}
}

func generateKustomizeBase(t *testing.T) {
	generateBaseKustomization(t)
	if t.Failed() {
		return
	}
	addKustomizationBaseImages(t)
}

func generateBaseKustomization(t *testing.T) {
	kContents := `
# This file is automatically generated by integration tests (env.go), any changes will be lost
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- deployment.yaml
- service_account.yaml
`
	err := os.MkdirAll(OperatorKustomizeBase, os.ModePerm)
	if !assert.Nil(t, err, "unable to create temp build directories") {
		return
	}
	for _, f := range []string{"deployment.yaml", "service_account.yaml"} {
		sourcePath := filepath.Join(OperatorSourceManifests, f)
		destPath := filepath.Join(OperatorKustomizeBase, f)
		out, err := os.Create(destPath)
		if !assert.Nil(t, err, fmt.Sprintf("unable to create file %s", destPath)) {
			return
		}
		in, err := os.Open(sourcePath)
		if !assert.Nil(t, err, fmt.Sprintf("unable to open file %s", sourcePath)) {
			return
		}
		_, err = io.Copy(out, in)
		if !assert.Nil(t, err, fmt.Sprintf("unable to copy %s to %s", sourcePath, destPath)) {
			return
		}
		in.Close()
	}
	err = ioutil.WriteFile(filepath.Join(OperatorKustomizeBase, "kustomization.yaml"), []byte(kContents), os.ModePerm)
	if !assert.Nil(t, err, "unable to write base kustomization.yaml file") {
		return
	}
}

func addKustomizationBaseImages(t *testing.T) {
	opImg := os.Getenv(OperatorImageVar)
	if opImg == "" {
		t.Logf("%s env var not set, using default", OperatorImageVar)
		opImg = OperatorImageDefault
	}
	t.Logf("Using operator image %s", opImg)
	halyardImg := os.Getenv(HalyardImageVar)
	if halyardImg == "" {
		t.Logf("%s env var not set, using default", HalyardImageVar)
		halyardImg = HalyardImageDefault
	}
	t.Logf("Using halyard image %s", halyardImg)
	RunCommandAndAssert(fmt.Sprintf("cd %s && kustomize edit set image spinnaker-operator=%s", OperatorKustomizeBase, opImg), t)
	if t.Failed() {
		return
	}
	RunCommandAndAssert(fmt.Sprintf("cd %s && kustomize edit set image halyard=%s", OperatorKustomizeBase, halyardImg), t)
}

func (e *TestEnv) KubectlPrefix() string {
	return fmt.Sprintf("kubectl --kubeconfig=%s", e.KubeconfigPath)
}

func (e *TestEnv) Cleanup(t *testing.T) {
	e.DeleteOperator(t)
}

func InstallCrdsAndOperator(opNs, opKustPath string, t *testing.T) (e *TestEnv) {
	LogMainStep(t, "Installing CRDs and operator in namespace %s", opNs)
	e = NewEnv(opNs, opKustPath, t)
	if t.Failed() {
		return e
	}
	if !e.InstallCrds(t) {
		return
	}
	e.InstallOperator(t)
	LogMainStep(t, "CRDs and operator installed")
	return
}

func (e *TestEnv) InstallCrds(t *testing.T) bool {
	ApplyManifest("default", CRDManifests, e, t)
	RunCommandAndAssert(fmt.Sprintf("%s get spinsvc", e.KubectlPrefix()), t)
	RunCommandAndAssert(fmt.Sprintf("%s get spinnakeraccounts", e.KubectlPrefix()), t)
	return !t.Failed()
}

func (e *TestEnv) InstallOperator(t *testing.T) bool {
	ran, ok := operatorRunsByNamespace[e.Operator.Namespace]
	if ok && ran {
		t.Logf("Operator already installed")
		return true
	}
	operatorRunsByNamespace[e.Operator.Namespace] = true
	if !CreateNamespace(e.Operator.Namespace, e, t) {
		return !t.Failed()
	}
	if !ApplyKustomizeAndAssert(e.Operator.Namespace, e.Operator.KustomizationPath, e, t) {
		return !t.Failed()
	}
	return WaitForDeploymentToStabilize(e.Operator.Namespace, "spinnaker-operator", e, t)
}

func (e *TestEnv) DeleteOperator(t *testing.T) {
	t.Logf("Deleting operator...")
	DeleteNamespace(e.Operator.Namespace, e, t)
}

func (e *TestEnv) InstallSpinnaker(ns, name, kustPath string, t *testing.T) bool {
	LogMainStep(t, "Installing spinnaker in namespace %s", ns)
	if !CreateNamespace(ns, e, t) {
		return !t.Failed()
	}
	e.SpinDeckUrl, e.SpinGateUrl = DeploySpinnaker(ns, name, kustPath, e, t)
	if t.Failed() {
		return !t.Failed()
	}
	LogMainStep(t, "Spinnaker installed successfully")
	return !t.Failed()
}

func (e *TestEnv) VerifyAccountsExist(t *testing.T, accts ...Account) bool {
	LogMainStep(t, "Verifying spinnaker accounts")
	o := ExecuteGetRequest(fmt.Sprintf("%s/credentials", e.SpinGateUrl), t)
	if t.Failed() {
		return !t.Failed()
	}
	var credentials []Account
	found := 0
	if assert.Nil(t, json.Unmarshal([]byte(o), &credentials)) {
		for _, a := range accts {
			for _, c := range credentials {
				if a.Type == c.Type && a.Name == c.Name {
					found++
					break
				}
			}
		}
	}
	assert.Equal(t, len(accts), found, fmt.Sprintf("Unable to find all accounts in spinnaker. Expected: %v but found: %v", accts, credentials))
	return !t.Failed()
}
