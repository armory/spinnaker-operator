package integration_tests

import (
	"fmt"
	"strings"
)

// TestEnv holds information about the kubernetes cluster used for tests
type TestEnv struct {
	KubeconfigPath string
	CRDpath        string
	Operator       Operator
}

// Operator holds information about the operator installation
type Operator struct {
	ManifestsPath string
	Namespace     string
}

func (e *TestEnv) KubectlPrefix() string {
	return fmt.Sprintf("kubectl --kubeconfig=%s ", e.KubeconfigPath)
}

func (e *TestEnv) Cleanup() {
	e.DeleteOperator()
}

func (e *TestEnv) InstallCrds() (string, error) {
	o, err := ApplyManifestWithError(e.CRDpath, e)
	// sometimes installing CRDs fails with this error
	if strings.Contains(o, "AlreadyExists") {
		return "", nil
	}
	return o, err
}

func (e *TestEnv) InstallOperator() (string, error) {
	println("Installing operator...")
	if o, err := CreateNamespaceWithError(e.Operator.Namespace, e); err != nil {
		return o, err
	}
	if o, err := ApplyManifestInNsWithError(e.Operator.ManifestsPath, e.Operator.Namespace, e); err != nil {
		return o, err
	}
	return WaitForManifestInNsToStabilizeWithError("pods", "operator", e.Operator.Namespace, e)
}

func (e *TestEnv) DeleteOperator() (string, error) {
	println("Deleting operator...")
	if o, err := DeleteManifestInNsWithError(e.Operator.ManifestsPath, e.Operator.Namespace, e); err != nil {
		return o, err
	}
	return DeleteNamespaceWithError(e.Operator.Namespace, e)
}
