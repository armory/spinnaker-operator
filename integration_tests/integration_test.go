// +build integration

package integration_tests

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

// - Operator in basic mode
// - Minimal spinnaker manifest (exposed)
func TestSpinnakerBase(t *testing.T) {
	// setup
	t.Parallel()
	e := InstallCrdsAndOperator(false, t)
	if t.Failed() {
		return
	}

	// install
	e.InstallSpinnaker(e.Operator.Namespace, "testdata/spinnaker/base", t)
}

// - Operator in cluster mode
// - Spinnaker with kubernetes accounts
// - Upgrade spinnaker
// - Uninstall spinnaker
func TestKubernetesAndUpgradeOverlay(t *testing.T) {
	// setup
	t.Parallel()
	spinOverlay := "testdata/spinnaker/overlay_kubernetes"
	ns := RandomString("spin-kubernetes-test")
	e := InstallCrdsAndOperator(true, t)
	if t.Failed() {
		return
	}

	// prepare overlay dynamic files
	LogMainStep(t, "Preparing overlay dynamic files for namespace %s", ns)
	if !e.GenerateSpinnakerRoleBinding(ns, spinOverlay, t) {
		return
	}

	// install
	if !e.InstallSpinnaker(ns, spinOverlay, t) {
		return
	}

	// verify accounts
	e.VerifyAccountsExist(t, Account{Name: "kube-sa-inline", Type: "kubernetes"})
	if t.Failed() {
		return
	}

	// upgrade
	LogMainStep(t, "Upgrading spinnaker")
	v := RunCommandAndAssert(fmt.Sprintf("%s -n %s get spinsvc %s -o=jsonpath='{.status.version}'", e.KubectlPrefix(), ns, SpinServiceName), t)
	if t.Failed() || !assert.Equal(t, "1.17.0", strings.TrimSpace(v)) {
		return
	}
	if !e.InstallSpinnaker(ns, "testdata/spinnaker/overlay_upgrade", t) {
		return
	}
	v = RunCommandAndAssert(fmt.Sprintf("%s -n %s get spinsvc %s -o=jsonpath='{.status.version}'", e.KubectlPrefix(), ns, SpinServiceName), t)
	if t.Failed() || !assert.Equal(t, "1.17.1", strings.TrimSpace(v)) {
		return
	}
	LogMainStep(t, "Upgrade successful")

	// uninstall
	LogMainStep(t, "Uninstalling spinnaker")
	RunCommandAndAssert(fmt.Sprintf("%s -n %s delete spinsvc %s", e.KubectlPrefix(), ns, SpinServiceName), t)
}
