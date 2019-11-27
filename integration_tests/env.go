package integration_tests

import (
	"fmt"
)

// Shared variable for all tests
var Env *TestEnv

// TestEnv holds information about the kubernetes cluster used for tests
type TestEnv struct {
	KubeconfigPath string
	CRDpath        string
}

func (e *TestEnv) KubectlPrefix() string {
	return fmt.Sprintf("kubectl --kubeconfig=%s ", e.KubeconfigPath)
}

func (e *TestEnv) Cleanup() {
	o, err := DeleteManifestWithError(e.CRDpath, e)
	if err != nil {
		println(fmt.Sprintf("Error deleting CRDs from cluster: %s, error: %v", o, err))
	}
}

func (e *TestEnv) InstallCrds() (string, error) {
	return ApplyManifestWithError(e.CRDpath, e)
}
