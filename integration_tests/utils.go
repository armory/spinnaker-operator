package integration_tests

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/stretchr/testify/assert"
)

const (
	SpinServiceName                        = "spinnaker"
	MaxErrorsWaitingForStability           = 3
	MaxChecksWaitingForDeploymentStability = 120 // (120 * 2s) / 60 = 4 minutes (large images may need to be downloaded + startup time)
	MaxChecksWaitingForSpinnakerStability  = 690 // (690 * 2s) / 60 = 23 minutes
	MaxChecksWaitingForLBStability         = 450 // (300 * 2s) / 60 = 15 minutes
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
	gateUrl = WaitForLBReady(ns, "{.status.apiUrl}", e, t)
	if t.Failed() {
		return
	}
	deckUrl = WaitForLBReady(ns, "{.status.uiUrl}", e, t)
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

func WaitForLBReady(ns, statusPath string, e *TestEnv, t *testing.T) string {
	t.Logf("Waiting for spinnaker lb (%s) to become reachable", statusPath)
	lbUrl := ""
	for counter := 0; counter < MaxChecksWaitingForLBStability; counter++ {
		if lbUrl == "" {
			lbUrl, _ = RunCommandSilent(fmt.Sprintf("%s -n %s get spinsvc %s -o=jsonpath='%s'", e.KubectlPrefix(), ns, SpinServiceName, statusPath), t)
		}
		if lbUrl != "" {
			_, err := RunCommandSilent(fmt.Sprintf("docker run --rm --network host curlimages/curl %s", lbUrl), t)
			if err == nil {
				return lbUrl
			}
		}
		time.Sleep(2 * time.Second)
	}
	o, _ := RunCommandSilent(fmt.Sprintf("%s -n %s get services", e.KubectlPrefix(), ns), t)
	t.Errorf("Waited too much time for spinnaker deck and gate LB's to be reachable. Either they're not assigned public LBs yet or DNS servers still don't resolve them. Services:\n%s", o)
	return ""
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

// ExecuteGetRequest
// Since the integration tests are interacting with the spinnaker URL and the URL is generated within the docker network.
// We need to emulate a curl call inside of the docker network, so we are using a docker image instead of the CURL command itself for it.
func ExecuteGetRequest(reqUrl string, t *testing.T) string {

	resp, err := RunCommand(fmt.Sprintf("docker run --rm --network host curlimages/curl -s %s", reqUrl), t)
	if !assert.Nil(t, err, fmt.Sprintf("Network error executing GET request to %s", reqUrl)) {
		t.Logf("GET request to %s failed with error: %s", reqUrl, err.Error())
		return ""
	}
	o := strings.TrimSpace(resp)
	assert.Nil(t, err, fmt.Sprintf("GET %s failed: %s", reqUrl, o))

	return o
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

func RunCommandInOperatorAndAssert(c string, e *TestEnv, t *testing.T) bool {
	RunCommandAndAssert(fmt.Sprintf("%s -n %s exec -c spinnaker-operator %s -- bash -c \"%s\"",
		e.KubectlPrefix(), e.Operator.Namespace, e.Operator.PodName, c), t)
	return !t.Failed()
}

func RunCommandInContainerAndAssert(ns, svc, cmd string, e *TestEnv, t *testing.T) string {
	pod := GetPodName(ns, svc, e, t)
	return RunCommandAndAssert(fmt.Sprintf("%s -n %s exec %s -- bash -c \"%s\"", e.KubectlPrefix(), ns, pod, cmd), t)
}

func CopyFileToS3Bucket(f, dest string, e *TestEnv, t *testing.T) bool {
	UpdateControlPlaneHost(f, t)
	RunCommandAndAssert(fmt.Sprintf("%s -n %s cp %s %s:/tmp/fileToCopy", e.KubectlPrefix(), e.Operator.Namespace, f, e.Operator.PodName), t)
	if t.Failed() {
		return !t.Failed()
	}
	c := fmt.Sprintf("aws s3 mv /tmp/fileToCopy s3://%s/%s", e.Vars.S3Bucket, dest)
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

func GetPodName(ns, svc string, e *TestEnv, t *testing.T) string {
	return strings.TrimSpace(RunCommandAndAssert(fmt.Sprintf("%s -n %s get pods | grep %s | grep \"1/1\" | grep \"Running\" | awk '{print $1}'", e.KubectlPrefix(), ns, svc), t))
}

func ExponentialBackOff(operation backoff.Operation, minutes time.Duration) error {
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = minutes * time.Minute
	b.MaxInterval = 20 * time.Minute

	return backoff.Retry(operation, b)
}

func GetLocalHost(t *testing.T) string {
	host, _ := RunCommandSilent("docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' minikube", t)
	return fmt.Sprintf(`https://%s:6443`, strings.TrimSpace(host))
}

func UpdateControlPlaneHost(path string, t *testing.T) {
	c, _ := RunCommandSilent(fmt.Sprintf("cat %s", path), t)
	re := regexp.MustCompile(`(http|https):\/\/([\w\-_]+(?:(?:\.[\w\-_]+)+))([\w\-\.,@?^=%&amp;:/~\+#]*[\w\-\@?^=%&amp;/~\+#])?`)
	c = re.ReplaceAllString(c, GetLocalHost(t))
	err := ioutil.WriteFile(path, []byte(c), os.ModePerm)
	assert.Nil(t, err, "unable to generate files.yml file")
}
