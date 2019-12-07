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
	LogMainStep(t, `Test goals:
- Install spinnaker with operator running in basic mode`)
	e := InstallCrdsAndOperator("", false, t)
	if t.Failed() {
		return
	}

	// install
	e.InstallSpinnaker(e.Vars.OperatorNamespace, "testdata/spinnaker/base", t)
}

func TestKubernetesAndUpgradeOverlay(t *testing.T) {
	// setup
	t.Parallel()
	LogMainStep(t, `Test goals:
- Install spinnaker with Kubernetes accounts:
  * Auth with service account
  * Auth with kubeconfigFile referencing a file inside inside spinConfig.files
- Upgrade spinnaker to a newer version
- Uninstall with kubectl delete spinsvc <name>`)

	spinOverlay := "testdata/spinnaker/overlay_kubernetes"
	ns := RandomString("spin-kubernetes-test")
	e := InstallCrdsAndOperator(ns, true, t)
	if t.Failed() {
		return
	}

	// prepare overlay dynamic files
	LogMainStep(t, "Preparing overlay dynamic files for namespace %s", ns)
	e.SubstituteOverlayVars(spinOverlay, t)
	if t.Failed() {
		return
	}
	if !e.GenerateSpinFiles(spinOverlay, "kubecfg", e.Vars.Kubeconfig, t) {
		return
	}
	defer RunCommand(fmt.Sprintf("rm %s/files.yml", spinOverlay), t)

	// install
	if !e.InstallSpinnaker(ns, spinOverlay, t) {
		return
	}

	// verify accounts
	e.VerifyAccountsExist(t,
		Account{Name: "kube-sa", Type: "kubernetes"},
		Account{Name: "kube-file-reference", Type: "kubernetes"})
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

func TestSecretsOverlay(t *testing.T) {
	// setup
	t.Parallel()
	LogMainStep(t, `Test goals:
- Install spinnaker with:
  * S3 secret values
  * S3 secret files
  * Kubernetes secret values
  * Kubernetes secret files
`)

	spinOverlay := "testdata/spinnaker/overlay_secrets"
	ns := RandomString("spin-secrets-test")
	e := InstallCrdsAndOperator(ns, true, t)
	if t.Failed() {
		return
	}

	// prepare overlay dynamic files
	LogMainStep(t, "Preparing overlay dynamic files for namespace %s", ns)
	RunCommandAndAssert(fmt.Sprintf("cp %s %s/kubecfg", e.Vars.Kubeconfig, spinOverlay), t)
	defer RunCommand(fmt.Sprintf("rm %s/kubecfg", spinOverlay), t)
	if t.Failed() {
		return
	}
	e.SubstituteOverlayVars(spinOverlay, t)
	if t.Failed() {
		return
	}

	// install
	e.InstallSpinnaker(ns, spinOverlay, t)
}
