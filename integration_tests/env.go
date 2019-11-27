package integration_tests

import (
	"fmt"
	"os"
	"os/exec"
)

// Shared variable for all tests
var Env *TestEnv

// TestEnv holds information about the kubernetes cluster used for tests
type TestEnv struct {
	VaultKey        string
	VaultPath       string
	LocalKubeconfig string
	CRDpath         string
}

func (e *TestEnv) KubectlPrefix() string {
	return fmt.Sprintf("kubectl --kubeconfig=%s ", e.LocalKubeconfig)
}

func (e *TestEnv) LoadKubeconfig() (string, error) {
	o, err := exec.Command("sh", "-c", fmt.Sprintf(
		"vault kv get -field %s %s > %s", e.VaultKey, e.VaultPath, e.LocalKubeconfig)).CombinedOutput()
	return string(o), err
}

func (e *TestEnv) Cleanup() {
	o, err := DeleteManifestWithError(e.CRDpath, e)
	if err != nil {
		println(fmt.Sprintf("Error deleting CRDs from cluster: %s, error: %v", o, err))
	}
	os.Remove(e.LocalKubeconfig)
}

func (e *TestEnv) InstallCrds() (string, error) {
	return ApplyManifestWithError(e.CRDpath, e)
}
