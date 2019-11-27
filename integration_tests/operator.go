package integration_tests

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type OperatorMode string

const (
	OperatorModeBasic   OperatorMode = "basic"
	OperatorModeCluster OperatorMode = "cluster"
)

// Operator holds information about the operator installation
type Operator struct {
	Env           *TestEnv
	OperatorMode  OperatorMode
	ManifestsPath string
	Namespace     string
}

func (o *Operator) InstallOperator(t *testing.T) {
	println("Installing operator...")
	if !assert.NotEqual(t, "", o.Namespace, "operator namespace is needed") {
		t.Fail()
		return
	}
	CreateNamespace(o.Namespace, o.Env, t)
	ApplyManifestInNamespace(o.ManifestsPath, o.Namespace, o.Env, t)
	// TODO: Create "wait for manifest to stabilize" helper functions
}

func (o *Operator) DeleteOperator(t *testing.T) {
	println("Deleting operator...")
	DeleteManifestInNamespace(o.ManifestsPath, o.Namespace, o.Env, t)
	DeleteNamespace(o.Namespace, o.Env, t)
}
