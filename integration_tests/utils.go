package integration_tests

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os/exec"
	"testing"
)

func ApplyManifest(path string, e *TestEnv, t *testing.T) {
	o, err := ApplyManifestWithError(path, e)
	if !assert.Nil(t, err, o) {
		t.FailNow()
	}
}

func ApplyManifestInNamespace(path, ns string, e *TestEnv, t *testing.T) {
	o, err := ApplyManifestInNamespaceWithError(path, ns, e)
	if !assert.Nil(t, err, o) {
		t.FailNow()
	}
}

func ApplyManifestWithError(path string, e *TestEnv) (string, error) {
	return ApplyManifestInNamespaceWithError(path, "", e)
}

func ApplyManifestInNamespaceWithError(path, ns string, e *TestEnv) (string, error) {
	normalizedNs := ns
	if normalizedNs == "" {
		normalizedNs = "default"
	}
	c := fmt.Sprintf("%s -n %s apply -f %s", e.KubectlPrefix(), normalizedNs, path)
	return RunCommand(c)
}

func DeleteManifest(path string, e *TestEnv, t *testing.T) {
	o, err := DeleteManifestWithError(path, e)
	if !assert.Nil(t, err, o) {
		t.FailNow()
	}
}

func DeleteManifestInNamespace(path, ns string, e *TestEnv, t *testing.T) {
	o, err := DeleteManifestInNamespaceWithError(path, ns, e)
	if !assert.Nil(t, err, o) {
		t.FailNow()
	}
}

func DeleteManifestWithError(path string, e *TestEnv) (string, error) {
	return DeleteManifestInNamespaceWithError(path, "", e)
}

func DeleteManifestInNamespaceWithError(path, ns string, e *TestEnv) (string, error) {
	normalizedNs := ns
	if normalizedNs == "" {
		normalizedNs = "default"
	}
	c := fmt.Sprintf("%s -n %s delete -f %s", e.KubectlPrefix(), normalizedNs, path)
	return RunCommand(c)
}

func CreateNamespace(name string, e *TestEnv, t *testing.T) {
	c := fmt.Sprintf("%s get namespace %s", e.KubectlPrefix(), name)
	o, err := RunCommand(c)
	if err == nil {
		// namespace already exists
		return
	}
	c = fmt.Sprintf("%s create namespace %s", e.KubectlPrefix(), name)
	o, err = RunCommand(c)
	if !assert.Nil(t, err, o) {
		t.FailNow()
	}
}

func DeleteNamespace(name string, e *TestEnv, t *testing.T) {
	c := fmt.Sprintf("%s delete namespace %s", e.KubectlPrefix(), name)
	o, err := RunCommand(c)
	if !assert.Nil(t, err, o) {
		t.FailNow()
	}
}

func RunCommand(c string) (string, error) {
	println(c)
	o, err := exec.Command("sh", "-c", c).CombinedOutput()
	s := string(o)
	println(s)
	return s, err
}
