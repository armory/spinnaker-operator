// +build integration

package integration_tests

import (
	"fmt"
	"os"
	"testing"
)

// TestMain is the entry point for all tests
func TestMain(m *testing.M) {
	Env = &TestEnv{
		VaultPath:       "secret/integration-tests/operator",
		VaultKey:        "kubeconfig",
		LocalKubeconfig: "int-test-kubeconfig",
		CRDpath:         "../deploy/crds",
	}
	println("Grabbing kubeconfig from vault")
	if o, err := Env.LoadKubeconfig(); err != nil {
		println(fmt.Sprintf("Error loading kubeconfig from vault: %s, error: %v", o, err))
	}
	println("Installing CRDs")
	if o, err := Env.InstallCrds(); err != nil {
		println(fmt.Sprintf("Error installing CRDs: %s, error: %v", o, err))
	}
	code := m.Run()
	Env.Cleanup()
	os.Exit(code)
}

func TestInstallSpinnaker(t *testing.T) {
	o := &Operator{
		Env:           Env,
		OperatorMode:  OperatorModeBasic,
		ManifestsPath: "../deploy/operator/basic",
		Namespace:     "operator",
	}
	o.InstallOperator(t)
	defer o.DeleteOperator(t)

	println("Actual spinnaker install happens here...")
}
