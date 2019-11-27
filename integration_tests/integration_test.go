// +build integration

package integration_tests

import (
	"fmt"
	"os"
	"testing"
)

// TestMain is the entry point for all tests
func TestMain(m *testing.M) {
	k := os.Getenv("KUBECONFIG")
	if k == "" {
		println("KUBECONFIG variable not set, falling back to $HOME/.kube/config")
		home, err := os.UserHomeDir()
		if err != nil {
			println("Error getting user home: %v", err)
			os.Exit(1)
		}
		k = fmt.Sprintf("%s/.kube/config", home)
	}
	println(fmt.Sprintf("Using kubeconfig %s", k))
	Env = &TestEnv{
		KubeconfigPath: k,
		CRDpath:        "../deploy/crds",
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
