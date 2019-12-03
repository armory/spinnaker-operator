package integration_tests

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func SetupEnv(crdPath, operatorPath, ns string) (*TestEnv, error) {
	k := os.Getenv("KUBECONFIG")
	if k == "" {
		println("KUBECONFIG variable not set, falling back to $HOME/.kube/config")
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("error getting user home: %v", err)
		}
		k = fmt.Sprintf("%s/.kube/config", home)
	}
	println(fmt.Sprintf("Using kubeconfig %s", k))
	e := &TestEnv{
		KubeconfigPath: k,
		CRDpath:        crdPath,
		Operator: Operator{
			ManifestsPath: operatorPath,
			Namespace:     ns,
		},
	}
	println("Installing CRDs")
	if o, err := e.InstallCrds(); err != nil {
		e.Cleanup()
		return nil, fmt.Errorf(fmt.Sprintf("Error installing CRDs: %s, error: %v", o, err))
	}
	if o, err := e.InstallOperator(); err != nil {
		e.Cleanup()
		return nil, fmt.Errorf(fmt.Sprintf("Error installing operator %s, error: %v", o, err))
	}
	return e, nil
}

// DeploySpinnaker returns spinnaker Deck and Gate public urls
func DeploySpinnaker(spinName, manifest, spinNs string, e *TestEnv, t *testing.T) (deckUrl string, gateUrl string) {
	o, err := ApplyManifestInNsWithError(manifest, spinNs, e)
	if err != nil {
		println(fmt.Sprintf("Error deploying spinnaker: %s, error: %v", o, err))
		PrintOperatorLogs(e)
		t.FailNow()
	}
	time.Sleep(3 * time.Second)
	WaitForSpinnakerToStabilize(spinName, spinNs, e, t)
	if t.Failed() {
		return "", ""
	}
	gateUrl, err = RunCommandSilent(fmt.Sprintf("%s -n %s get spinsvc %s -o=jsonpath='{.status.apiUrl}'", e.KubectlPrefix(), spinNs, spinName))
	if err != nil {
		println(fmt.Sprintf("Cannot get Gate public url: %s, error: %v", gateUrl, err))
		t.FailNow()
	}
	deckUrl, err = RunCommandSilent(fmt.Sprintf("%s -n %s get spinsvc %s -o=jsonpath='{.status.uiUrl}'", e.KubectlPrefix(), spinNs, spinName))
	if err != nil {
		println(fmt.Sprintf("Cannot get Deck public url: %s, error: %v", deckUrl, err))
		t.FailNow()
	}
	return
}

func PrintOperatorLogs(e *TestEnv) {
	o, _ := RunCommandSilent(fmt.Sprintf("%s -n %s logs deployment/spinnaker-operator spinnaker-operator", e.KubectlPrefix(), e.Operator.Namespace))
	println("================ Operator logs start ================ ")
	println(o)
	println("================ Operator logs end ================== ")
}

func ApplyManifest(path string, e *TestEnv, t *testing.T) {
	o, err := ApplyManifestWithError(path, e)
	if !assert.Nil(t, err, o) {
		t.FailNow()
	}
}

func ApplyManifestInNs(path, ns string, e *TestEnv, t *testing.T) {
	o, err := ApplyManifestInNsWithError(path, ns, e)
	if !assert.Nil(t, err, o) {
		t.FailNow()
	}
}

func ApplyManifestWithError(path string, e *TestEnv) (string, error) {
	return ApplyManifestInNsWithError(path, "", e)
}

func ApplyManifestInNsWithError(path, ns string, e *TestEnv) (string, error) {
	normalizedNs := ns
	if normalizedNs == "" {
		normalizedNs = "default"
	}
	c := fmt.Sprintf("%s -n %s apply -f %s", e.KubectlPrefix(), normalizedNs, path)
	return RunCommand(c)
}

func WaitForSpinnakerToStabilize(spinName, ns string, e *TestEnv, t *testing.T) {
	c := fmt.Sprintf("%s -n %s get spinsvc %s -o=jsonpath='{.status.status}'", e.KubectlPrefix(), ns, spinName)
	println(fmt.Sprintf("Waiting for spinnaker to become ready (%s)", c))
	errorCounter := 0
	for counter := 0; counter < 150; counter++ {
		print(".")
		o, err := RunCommandSilent(c)
		if err != nil {
			// fail only in repeated failures of "kubectl get spinsvc" command to avoid sporadic comms errors
			errorCounter++
			if errorCounter > 3 {
				assert.Fail(t, fmt.Sprintf("Error waiting for spinnaker to become ready: %s", o))
				return
			}
		}
		if strings.TrimSpace(o) == "OK" {
			println("\n")
			return
		}
		time.Sleep(2 * time.Second)
	}
	o, _ := RunCommandSilent(fmt.Sprintf("%s -n %s get pods", e.KubectlPrefix(), ns))
	assert.Fail(t, fmt.Sprintf("\nWaited too much time for spinnaker to become ready. Pods:\n%s", o))
}

func WaitForManifestInNsToStabilize(kind, resName, ns string, e *TestEnv, t *testing.T) {
	o, err := WaitForManifestInNsToStabilizeWithError(kind, resName, ns, e)
	if !assert.Nil(t, err, o) {
		t.FailNow()
	}
}

func WaitForManifestInNsToStabilizeWithError(kind, resName, ns string, e *TestEnv) (string, error) {
	normalizedNs := ns
	if normalizedNs == "" {
		normalizedNs = "default"
	}
	c := fmt.Sprintf("%s -n %s get %s | grep %s | awk '{print $2}'", e.KubectlPrefix(), normalizedNs, kind, resName)
	println(fmt.Sprintf("Waiting for manifest to stabilize (%s)", c))

	for counter := 0; counter < 20; counter++ {
		print(".")
		cont, err := RunCommandSilent(c)
		if err != nil {
			return cont, err
		}
		parts := strings.Split(cont, "/")
		if len(parts) == 2 && strings.TrimSpace(parts[0]) == strings.TrimSpace(parts[1]) {
			println("\n")
			return "", nil
		}
		time.Sleep(2 * time.Second)
	}
	return "", fmt.Errorf("\nwaited too much for %s to become ready, giving up", resName)
}

func DeleteManifest(path string, e *TestEnv, t *testing.T) {
	o, err := DeleteManifestWithError(path, e)
	if !assert.Nil(t, err, o) {
		t.FailNow()
	}
}

func DeleteManifestInNs(path, ns string, e *TestEnv, t *testing.T) {
	o, err := DeleteManifestInNsWithError(path, ns, e)
	if !assert.Nil(t, err, o) {
		t.FailNow()
	}
}

func DeleteManifestWithError(path string, e *TestEnv) (string, error) {
	return DeleteManifestInNsWithError(path, "", e)
}

func DeleteManifestInNsWithError(path, ns string, e *TestEnv) (string, error) {
	normalizedNs := ns
	if normalizedNs == "" {
		normalizedNs = "default"
	}
	c := fmt.Sprintf("%s -n %s delete -f %s", e.KubectlPrefix(), normalizedNs, path)
	return RunCommand(c)
}

func CreateNamespace(name string, e *TestEnv, t *testing.T) {
	o, err := CreateNamespaceWithError(name, e)
	if !assert.Nil(t, err, o) {
		t.FailNow()
	}
}

func CreateNamespaceWithError(name string, e *TestEnv) (string, error) {
	c := fmt.Sprintf("%s get namespace %s", e.KubectlPrefix(), name)
	_, err := RunCommand(c)
	if err == nil {
		// namespace already exists
		return "", nil
	}
	c = fmt.Sprintf("%s create namespace %s", e.KubectlPrefix(), name)
	return RunCommand(c)
}

func DeleteNamespace(name string, e *TestEnv, t *testing.T) {
	o, err := DeleteNamespaceWithError(name, e)
	if !assert.Nil(t, err, o) {
		t.FailNow()
	}
}

func DeleteNamespaceWithError(name string, e *TestEnv) (string, error) {
	c := fmt.Sprintf("%s delete namespace %s", e.KubectlPrefix(), name)
	return RunCommand(c)
}

func RunCommand(c string) (string, error) {
	println(c)
	o, err := exec.Command("sh", "-c", c).CombinedOutput()
	s := string(o)
	println(s)
	return s, err
}

func RunCommandSilent(c string) (string, error) {
	o, err := exec.Command("sh", "-c", c).CombinedOutput()
	s := string(o)
	return s, err
}

func ExecuteGetRequest(reqUrl string, t *testing.T) string {
	req, err := http.NewRequest("GET", reqUrl, nil)
	if assert.Nil(t, err) {
		req = req.WithContext(context.TODO())
		client := &http.Client{}
		resp, err := client.Do(req)
		if assert.Nil(t, err) {
			defer resp.Body.Close()
			b, _ := ioutil.ReadAll(resp.Body)
			return string(b)
		}
	}
	return ""
}
