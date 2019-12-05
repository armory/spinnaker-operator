// +build integration

package integration_tests

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

const (
	OperatorClusterNs = "test-operator-cluster-mode" // Needs to be a constant because RoleBinding manifest references it
	OperatorBasicNs   = "test-spinnaker-basic-mode"
)

func TestOperatorBasicMode(t *testing.T) {
	// setup
	t.Parallel()
	e := InstallCrdsAndOperator(OperatorBasicNs, "testdata/operator/overlay_basicmode", t)
	if t.Failed() {
		return
	}

	// install
	spinName := "spinnaker"
	e.InstallSpinnaker(OperatorBasicNs, spinName, "testdata/spinnaker/base", t)
}

func TestInstallUpgradeUninstall(t *testing.T) {
	// setup
	t.Parallel()
	e := InstallCrdsAndOperator(OperatorClusterNs, "testdata/operator/overlay_clustermode", t)
	if t.Failed() {
		return
	}

	// install
	ns := "test-spinnaker-cluster-mode"
	spinName := "spinnaker"
	if !e.InstallSpinnaker(ns, spinName, "testdata/spinnaker/overlay_kubernetes", t) {
		return
	}

	// verify accounts
	e.VerifyAccountsExist(t, Account{Name: "kube-sa-inline", Type: "kubernetes"})
	if t.Failed() {
		return
	}

	// upgrade
	LogMainStep(t, "Upgrading spinnaker")
	v := RunCommandAndAssert(fmt.Sprintf("%s -n %s get spinsvc %s -o=jsonpath='{.status.version}'", e.KubectlPrefix(), ns, spinName), t)
	if t.Failed() || !assert.Equal(t, "1.17.0", strings.TrimSpace(v)) {
		return
	}
	if !e.InstallSpinnaker(ns, spinName, "testdata/spinnaker/overlay_upgrade", t) {
		return
	}
	v = RunCommandAndAssert(fmt.Sprintf("%s -n %s get spinsvc %s -o=jsonpath='{.status.version}'", e.KubectlPrefix(), ns, spinName), t)
	if t.Failed() || !assert.Equal(t, "1.17.1", strings.TrimSpace(v)) {
		return
	}
	LogMainStep(t, "Upgrade successful")

	// uninstall
	LogMainStep(t, "Uninstalling spinnaker")
	RunCommandAndAssert(fmt.Sprintf("%s -n %s delete spinsvc %s", e.KubectlPrefix(), ns, spinName), t)
}
