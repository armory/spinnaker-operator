// +build integration

package integration_tests

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

// - Operator in basic mode
// - Minimal spinnaker manifest
func TestSuccessfulInstall_1(t *testing.T) {
	// setup
	t.Parallel()
	e := InstallCrdsAndOperator(NsOperatorBasic, "testdata/operator/overlay_basicmode", t)
	if t.Failed() {
		return
	}

	// install
	spinName := "spinnaker"
	e.InstallSpinnaker(NsOperatorBasic, spinName, "testdata/spinnaker/base", t)
}

// - Operator in cluster mode
// - Spinnaker with kubernetes accounts
// - Upgrade spinnaker
// - Uninstall spinnaker
func TestSuccessfulInstall_2(t *testing.T) {
	// setup
	t.Parallel()
	e := InstallCrdsAndOperator(NsOperatorCluster, "testdata/operator/overlay_clustermode", t)
	if t.Failed() {
		return
	}

	// install
	spinName := "spinnaker"
	if !e.InstallSpinnaker(NsSpinnaker1, spinName, "testdata/spinnaker/overlay_kubernetes", t) {
		return
	}

	// verify accounts
	e.VerifyAccountsExist(t, Account{Name: "kube-sa-inline", Type: "kubernetes"})
	if t.Failed() {
		return
	}

	// upgrade
	LogMainStep(t, "Upgrading spinnaker")
	v := RunCommandAndAssert(fmt.Sprintf("%s -n %s get spinsvc %s -o=jsonpath='{.status.version}'", e.KubectlPrefix(), NsSpinnaker1, spinName), t)
	if t.Failed() || !assert.Equal(t, "1.17.0", strings.TrimSpace(v)) {
		return
	}
	if !e.InstallSpinnaker(NsSpinnaker1, spinName, "testdata/spinnaker/overlay_upgrade", t) {
		return
	}
	v = RunCommandAndAssert(fmt.Sprintf("%s -n %s get spinsvc %s -o=jsonpath='{.status.version}'", e.KubectlPrefix(), NsSpinnaker1, spinName), t)
	if t.Failed() || !assert.Equal(t, "1.17.1", strings.TrimSpace(v)) {
		return
	}
	LogMainStep(t, "Upgrade successful")

	// uninstall
	LogMainStep(t, "Uninstalling spinnaker")
	RunCommandAndAssert(fmt.Sprintf("%s -n %s delete spinsvc %s", e.KubectlPrefix(), NsSpinnaker1, spinName), t)
}
