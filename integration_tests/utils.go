package integration_tests

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"html/template"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	SpinServiceName                        = "spinnaker"
	MaxErrorsWaitingForStability           = 3
	MaxChecksWaitingForDeploymentStability = 25
	MaxChecksWaitingForSpinnakerStability  = 250
)

var SpinBaseSvcs []string

func init() {
	SpinBaseSvcs = append(SpinBaseSvcs, "spin-deck")
	SpinBaseSvcs = append(SpinBaseSvcs, "spin-gate")
	SpinBaseSvcs = append(SpinBaseSvcs, "spin-orca")
	SpinBaseSvcs = append(SpinBaseSvcs, "spin-clouddriver")
	SpinBaseSvcs = append(SpinBaseSvcs, "spin-echo")
	SpinBaseSvcs = append(SpinBaseSvcs, "spin-front50")
	SpinBaseSvcs = append(SpinBaseSvcs, "spin-rosco")
}

// DeploySpinnaker returns spinnaker Deck and Gate public urls
func DeploySpinnaker(ns, kustPath string, e *TestEnv, t *testing.T) (deckUrl string, gateUrl string) {
	if !ApplyKustomizeAndAssert(ns, kustPath, e, t) {
		t.Logf("Error deploying spinnaker")
		PrintOperatorLogs(e, t)
		return
	}
	time.Sleep(3 * time.Second)
	WaitForSpinnakerToStabilize(ns, e, t)
	if t.Failed() {
		return
	}
	gateUrl = RunCommandSilentAndAssert(fmt.Sprintf("%s -n %s get spinsvc %s -o=jsonpath='{.status.apiUrl}'", e.KubectlPrefix(), ns, SpinServiceName), t)
	if t.Failed() {
		t.Logf("Cannot get Gate public url: %s", gateUrl)
		return
	}
	deckUrl = RunCommandSilentAndAssert(fmt.Sprintf("%s -n %s get spinsvc %s -o=jsonpath='{.status.uiUrl}'", e.KubectlPrefix(), ns, SpinServiceName), t)
	if t.Failed() {
		t.Logf("Cannot get Deck public url: %s", deckUrl)
		return
	}
	return
}

func PrintOperatorLogs(e *TestEnv, t *testing.T) {
	o, _ := RunCommandSilent(fmt.Sprintf("%s -n %s logs deployment/spinnaker-operator spinnaker-operator", e.KubectlPrefix(), e.Operator.Namespace), t)
	t.Logf("================ Operator logs start ================ ")
	t.Logf(o)
	t.Logf("================ Operator logs end ================== ")
}

func ApplyManifest(ns, path string, e *TestEnv, t *testing.T) {
	c := fmt.Sprintf("%s -n %s apply -f %s", e.KubectlPrefix(), ns, path)
	RunCommand(c, t)
}

func ApplyKustomize(ns, path string, e *TestEnv, t *testing.T) (string, error) {
	c := fmt.Sprintf("%s -n %s apply -k %s", e.KubectlPrefix(), ns, path)
	return RunCommand(c, t)
}

func ApplyKustomizeAndAssert(ns, path string, e *TestEnv, t *testing.T) bool {
	c := fmt.Sprintf("%s -n %s apply -k %s", e.KubectlPrefix(), ns, path)
	RunCommandAndAssert(c, t)
	return !t.Failed()
}

func WaitForSpinnakerToStabilize(ns string, e *TestEnv, t *testing.T) {
	c := fmt.Sprintf("%s -n %s get spinsvc %s -o=jsonpath='{.status.status}'", e.KubectlPrefix(), ns, SpinServiceName)
	t.Logf("Waiting for spinnaker to become ready (%s)", c)
	errCount := 0
	for counter := 0; counter < MaxChecksWaitingForSpinnakerStability; counter++ {
		o, err := RunCommandSilent(c, t)
		if err != nil {
			errCount++
			if !assert.NotEqual(t, MaxErrorsWaitingForStability, errCount,
				fmt.Sprintf("Waiting for spinnaker to become ready produced too many errors. Last output: %s", o)) {
				return
			}
		}
		if strings.TrimSpace(o) == "OK" {
			AssertSpinnakerHealthy(ns, SpinServiceName, e, t)
			return
		}
		time.Sleep(2 * time.Second)
	}
	o, _ := RunCommandSilent(fmt.Sprintf("%s -n %s get pods", e.KubectlPrefix(), ns), t)
	t.Errorf("Waited too much time for spinnaker to become ready (never saw status=OK). Pods:\n%s", o)
}

func AssertSpinnakerHealthy(ns, spinName string, e *TestEnv, t *testing.T) {
	t.Logf("Asserting spinnaker pods are healthy")
	for _, s := range SpinBaseSvcs {
		o := RunCommandSilentAndAssert(fmt.Sprintf("%s -n %s get deployment/%s -o=jsonpath='{.status.readyReplicas}'", e.KubectlPrefix(), ns, s), t)
		if t.Failed() {
			return
		}
		if !assert.Equal(t, "1", strings.TrimSpace(o), fmt.Sprintf("Expected %s deployment to have %d ready replicas, but was %s", s, 1, o)) {
			return
		}
	}
}

func WaitForDeploymentToStabilize(ns, name string, e *TestEnv, t *testing.T) bool {
	c := fmt.Sprintf("%s -n %s get deployment %s -o=jsonpath='{.status.updatedReplicas}/{.status.replicas}/{.status.unavailableReplicas}'", e.KubectlPrefix(), ns, name)
	t.Logf("Waiting for deployment %s to stabilize (command %s)", name, c)
	errCount := 0
	for counter := 0; counter < MaxChecksWaitingForDeploymentStability; counter++ {
		cont, err := RunCommandSilent(c, t)
		if err != nil {
			errCount++
			if !assert.NotEqual(t, MaxErrorsWaitingForStability, errCount,
				fmt.Sprintf("waiting for deployment %s to become ready produced too many errors. Last output: %s", name, cont)) {
				return !t.Failed()
			}
		}
		parts := strings.Split(cont, "/")
		if len(parts) == 3 && strings.TrimSpace(parts[0]) == strings.TrimSpace(parts[1]) && strings.TrimSpace(parts[2]) == "" {
			return !t.Failed()
		}
		time.Sleep(2 * time.Second)
	}
	pods, _ := RunCommandSilent(fmt.Sprintf("%s -n %s get pods", e.KubectlPrefix(), ns), t)
	t.Errorf("Waited too much for deployment %s to become ready, giving up. Pods: \n%s", name, pods)
	return !t.Failed()
}

func CreateNamespace(name string, e *TestEnv, t *testing.T) bool {
	RunCommandAndAssert(fmt.Sprintf("%s get ns %s || %s create ns %s", e.KubectlPrefix(), name, e.KubectlPrefix(), name), t)
	return !t.Failed()
}

func DeleteNamespace(name string, e *TestEnv, t *testing.T) {
	c := fmt.Sprintf("%s delete namespace %s", e.KubectlPrefix(), name)
	RunCommand(c, t)
}

func RunCommandSilent(c string, t *testing.T) (string, error) {
	o, err := exec.Command("sh", "-c", c).CombinedOutput()
	s := string(o)
	return s, err
}

func RunCommand(c string, t *testing.T) (string, error) {
	t.Logf("%s", c)
	o, err := exec.Command("sh", "-c", c).CombinedOutput()
	s := string(o)
	t.Logf("%s", s)
	return s, err
}

func RunCommandAndAssert(c string, t *testing.T) string {
	t.Logf("%s", c)
	o, err := exec.Command("sh", "-c", c).CombinedOutput()
	assert.Nil(t, err, fmt.Sprintf("command \"%s\" failed. Output: %s", c, o))
	s := string(o)
	t.Logf("%s", s)
	return s
}

func RunCommandSilentAndAssert(c string, t *testing.T) string {
	o, err := exec.Command("sh", "-c", c).CombinedOutput()
	assert.Nil(t, err, fmt.Sprintf("command \"%s\" failed. Output: %s", c, o))
	s := string(o)
	return s
}

func ExecuteGetRequest(reqUrl string, t *testing.T) string {
	req, err := http.NewRequest("GET", reqUrl, nil)
	if assert.Nil(t, err) {
		req = req.WithContext(context.TODO())
		client := &http.Client{}
		resp, err := client.Do(req)
		defer resp.Body.Close()
		b, _ := ioutil.ReadAll(resp.Body)
		o := string(b)
		assert.Nil(t, err, fmt.Sprintf("GET %s failed: %s", reqUrl, o))
		return o
	}
	return ""
}

func LogMainStep(t *testing.T, msg string, args ...interface{}) {
	if args == nil {
		t.Logf(fmt.Sprintf("================================ %s", msg))
	} else {
		t.Logf(fmt.Sprintf("================================ %s", msg), args)
	}
}

func RandomString(prefix string) string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%s-%d", prefix, rand.Intn(999))
}

func InstallAwsCli(e *TestEnv, t *testing.T) bool {
	c := "wget -O /tmp/get-pip.py https://bootstrap.pypa.io/get-pip.py && python /tmp/get-pip.py --user && /home/spinnaker-operator/.local/bin/pip install --user --upgrade awscli==1.16.208"
	RunCommandSilentAndAssert(fmt.Sprintf("%s -n %s exec -c spinnaker-operator %s -- bash -c \"%s\"",
		e.KubectlPrefix(), e.Operator.Namespace, e.Operator.PodName, c), t)
	return !t.Failed()
}

func RunCommandInOperatorAndAssert(c string, e *TestEnv, t *testing.T) bool {
	RunCommandAndAssert(fmt.Sprintf("%s -n %s exec -c spinnaker-operator %s -- bash -c \"%s\"",
		e.KubectlPrefix(), e.Operator.Namespace, e.Operator.PodName, c), t)
	return !t.Failed()
}

func CopyFileToS3Bucket(f, dest string, e *TestEnv, t *testing.T) bool {
	RunCommandAndAssert(fmt.Sprintf("%s -n %s cp %s %s:/tmp/fileToCopy", e.KubectlPrefix(), e.Operator.Namespace, f, e.Operator.PodName), t)
	if t.Failed() {
		return !t.Failed()
	}
	c := fmt.Sprintf("/home/spinnaker-operator/.local/bin/aws s3 mv /tmp/fileToCopy s3://%s/%s", e.Vars.S3Bucket, dest)
	return RunCommandInOperatorAndAssert(c, e, t)
}

func SubstituteOverlayVars(overlayHome string, vars interface{}, t *testing.T) bool {
	fs, err := ioutil.ReadDir(overlayHome)
	if !assert.Nil(t, err) {
		return !t.Failed()
	}
	for _, f := range fs {
		if !strings.Contains(f.Name(), "-template") {
			continue
		}
		tmpl, err := template.New(f.Name()).ParseFiles(filepath.Join(overlayHome, f.Name()))
		if !assert.Nil(t, err) {
			return !t.Failed()
		}
		n := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
		n = strings.ReplaceAll(n, "-template", "")
		p := filepath.Join(overlayHome, fmt.Sprintf("%s-generated%s", n, filepath.Ext(f.Name())))
		gf, err := os.Create(p)
		if !assert.Nil(t, err) {
			return !t.Failed()
		}
		if !assert.Nil(t, tmpl.ExecuteTemplate(gf, f.Name(), vars)) {
			return !t.Failed()
		}
	}
	return !t.Failed()
}
