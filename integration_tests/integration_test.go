// +build integration

package integration_tests

import (
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/controller/spinnakerservice"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

var defaults Defaults

func init() {
	defaults = Defaults{
		OperatorImageDefault:  "armory/spinnaker-operator:dev",
		HalyardImageDefault:   "armory/halyard:operator-dev",
		BucketDefault:         "operator-int-tests",
		BucketRegionDefault:   "us-west-2",
		OperatorKustomizeBase: "testdata/operator/base",
		CRDManifests:          "../deploy/crds",
	}
}

func TestSpinnakerBase(t *testing.T) {
	// setup
	t.Parallel()
	LogMainStep(t, `Test goals:
- Install spinnaker with operator running in basic mode`)
	e := InstallCrdsAndOperator("", false, defaults, t)
	if t.Failed() {
		return
	}

	// install
	e.InstallSpinnaker(e.Operator.Namespace, "testdata/spinnaker/base", t)
}

func TestKubernetesAndUpgradeOverlay(t *testing.T) {
	// setup
	t.Parallel()
	LogMainStep(t, `Test goals:
- Install spinnaker with Kubernetes accounts:
  * Auth with service account
  * Auth with kubeconfigFile referencing a file inside inside spinConfig.files
- Install spinnaker with anonymous dockerhub account
- Upgrade spinnaker to a newer version
- Uninstall with kubectl delete spinsvc <name>`)

	spinOverlay := "testdata/spinnaker/overlay_kubernetes"
	ns := RandomString("spin-kubernetes-test")
	e := InstallCrdsAndOperator(ns, true, defaults, t)
	if t.Failed() {
		return
	}

	// prepare overlay dynamic files
	LogMainStep(t, "Preparing overlay dynamic files for namespace %s", ns)
	SubstituteOverlayVars(spinOverlay, e.Vars, t)
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
	e.VerifyAccountsExist("/credentials", t,
		Account{Name: "dockerhub", Type: "dockerRegistry"},
		Account{Name: "kube-sa", Type: "kubernetes"},
		Account{Name: "kube-file-reference", Type: "kubernetes"})
	if t.Failed() {
		return
	}

	// upgrade
	LogMainStep(t, "Upgrading spinnaker")
	v := RunCommandAndAssert(fmt.Sprintf("%s -n %s get spinsvc %s -o=jsonpath='{.status.version}'", e.KubectlPrefix(), ns, SpinServiceName), t)
	if t.Failed() || !assert.Equal(t, "1.20.5", strings.TrimSpace(v)) {
		return
	}
	if !e.InstallSpinnaker(ns, "testdata/spinnaker/overlay_upgrade", t) {
		return
	}
	time.Sleep(20 * time.Second)
	v = RunCommandAndAssert(fmt.Sprintf("%s -n %s get spinsvc %s -o=jsonpath='{.status.version}'", e.KubectlPrefix(), ns, SpinServiceName), t)
	if t.Failed() || !assert.Equal(t, "1.17.1", strings.TrimSpace(v)) {
		return
	}
	LogMainStep(t, "Upgrade successful")

	// uninstall
	LogMainStep(t, "Uninstalling spinnaker")
	RunCommandAndAssert(fmt.Sprintf("%s -n %s delete spinsvc %s", e.KubectlPrefix(), ns, SpinServiceName), t)
}

func TestUpdateSpinsvcStatus(t *testing.T) {
	// setup
	t.Parallel()
	LogMainStep(t, `Test goals:
- Install spinnaker
- Check an OK status
- Upgrade spinnaker with a bad configurations
- Check a Failure status`)

	spinOverlay := "testdata/spinnaker/base"
	ns := RandomString("spinsvc-status")
	e := InstallCrdsAndOperator(ns, true, defaults, t)
	if t.Failed() {
		return
	}

	// install
	if !e.InstallSpinnaker(ns, spinOverlay, t) {
		return
	}

	v := RunCommandAndAssert(fmt.Sprintf("%s -n %s get spinsvc %s -o=jsonpath='{.status.status}'", e.KubectlPrefix(), ns, SpinServiceName), t)
	if t.Failed() || !assert.Equal(t, spinnakerservice.Ok, strings.TrimSpace(v)) {
		return
	}

	if !e.InstallSpinnaker(ns, "testdata/spinnaker/overlay_spinsvc_status", t) {
		return
	}

	ExponentialBackOff(func(ns, status string, e *TestEnv, t *testing.T) error {

		v := RunCommandAndAssert(fmt.Sprintf("%s -n %s get spinsvc %s -o=jsonpath='{.status.status}'", e.KubectlPrefix(), ns, SpinServiceName), t)
		if t.Failed() || !assert.Equal(t, status, strings.TrimSpace(v)) {
			return fmt.Errorf("spinnaker is not in %s status yet", status)
		}

		return nil
	}(t, ns, spinnakerservice.Failure, e, t), 3)

	// uninstall
	LogMainStep(t, "Uninstalling spinnaker")
	RunCommandAndAssert(fmt.Sprintf("%s -n %s delete spinsvc %s", e.KubectlPrefix(), ns, SpinServiceName), t)
}

func TestSecretsAndDuplicateOverlay(t *testing.T) {
	// setup
	t.Parallel()
	LogMainStep(t, `Test goals:
- Install spinnaker with:
  * S3 secret values
  * S3 secret files
  * Kubernetes secret values
  * Kubernetes secret files
- Try to install a second spinnaker in the same namespace, should fail
`)

	spinOverlay := "testdata/spinnaker/overlay_secrets"
	ns := RandomString("spin-secrets-test")
	e := InstallCrdsAndOperator(ns, true, defaults, t)
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
	SubstituteOverlayVars(spinOverlay, e.Vars, t)
	if t.Failed() {
		return
	}

	// store needed secrets in S3 bucket
	CopyFileToS3Bucket(fmt.Sprintf("%s/kubecfg", spinOverlay), "secrets/kubeconfig", e, t)
	if !CopyFileToS3Bucket(fmt.Sprintf("%s/secrets.yml", spinOverlay), "secrets/secrets.yml", e, t) {
		return
	}

	// install
	if !e.InstallSpinnaker(ns, spinOverlay, t) {
		return
	}

	// verify accounts
	e.VerifyAccountsExist("/credentials", t,
		Account{Name: "kube-s3-secret", Type: "kubernetes"},
		Account{Name: "kube-k8s-secret", Type: "kubernetes"})
	e.VerifyAccountsExist("/artifacts/credentials", t,
		Account{Name: "test-github-account-s3", Types: []string{"github/file"}},
		Account{Name: "test-github-account-k8s", Types: []string{"github/file"}},
		Account{Name: "test-s3-account-1", Types: []string{"s3/object"}})
	if t.Failed() {
		return
	}

	// try to install a second spinnaker in the same namespace
	o, err := ApplyKustomize(e.Operator.Namespace, "testdata/spinnaker/overlay_duplicate", e, t)
	assert.NotNil(t, err, fmt.Sprintf("expected error but was %s", o))
}

func TestProfilesOverlay(t *testing.T) {
	// setup
	t.Parallel()
	LogMainStep(t, `Test goals:
- Settings in profile configs should work 
- service-settings are merged and applied
`)

	spinOverlay := "testdata/spinnaker/overlay_profiles"
	ns := RandomString("spin-profiles-test")
	e := InstallCrdsAndOperator(ns, true, defaults, t)
	if t.Failed() {
		return
	}

	SubstituteOverlayVars(spinOverlay, e.Vars, t)
	if t.Failed() {
		return
	}

	// install
	e.InstallSpinnaker(ns, spinOverlay, t)
	if t.Failed() {
		return
	}

	// validate
	e.VerifyAccountsExist("/credentials", t, Account{Name: "kube-sa", Type: "kubernetes"})
	o := ExecuteGetRequest(fmt.Sprintf("%s/settings-local.js", e.SpinDeckUrl), t)
	assert.True(t, strings.Contains(o, "window.spinnakerSettings.feature.kustomizeEnabled"))
	j := strings.TrimSpace(RunCommandInContainerAndAssert(ns, "spin-rosco", fmt.Sprintf("cat /opt/rosco/config/packer/example-packer-config.json"), e, t))
	assert.Equal(t,
		`{
  "key1": "value1",
  "key2": "value2"
}`, j)
	sh := strings.TrimSpace(RunCommandInContainerAndAssert(ns, "spin-rosco", fmt.Sprintf("cat /opt/rosco/config/packer/my_custom_script.sh"), e, t))
	assert.Equal(t,
		`#!/bin/bash -e
echo "hello world!"`, sh)
	// all services have the global var
	for _, svc := range []string{"clouddriver", "echo", "front50", "gate", "orca", "rosco"} {
		pod := GetPodName(ns, svc, e, t)
		c := fmt.Sprintf("%s -n %s get pod %s -o=jsonpath='{.spec.containers[0].env[?(@.name==\"GLOBAL_VAR\")]}'", e.KubectlPrefix(), ns, pod)
		o := RunCommandSilentAndAssert(c, t)
		assert.NotEqual(t, "", strings.TrimSpace(o))
	}
	// only clouddriver has the extra var
	pod := GetPodName(ns, "clouddriver", e, t)
	c := fmt.Sprintf("%s -n %s get pod %s -o=jsonpath='{.spec.containers[0].env[?(@.name==\"SVC_NAME\")]}'", e.KubectlPrefix(), ns, pod)
	o = RunCommandSilentAndAssert(c, t)
	assert.NotEqual(t, "", strings.TrimSpace(o))
}

func TestValidations(t *testing.T) {
	// setup
	t.Parallel()
	LogMainStep(t, `Test goals:
- Validations can be enabled/disabled `)

	spinOverlay := "testdata/spinnaker/overlay_validations"
	ns := RandomString("spin-validations-test")
	e := InstallCrdsAndOperator(ns, true, defaults, t)
	if t.Failed() {
		return
	}
	LogMainStep(t, "Installing spinnaker in namespace %s", ns)
	if !CreateNamespace(ns, e, t) {
		return
	}

	// Apply manifests with errors
	vars := map[string]bool{"PersistentS3Enabled": true, "KubernetesEnabled": true, "DockerEnabled": true}
	SubstituteOverlayVars(spinOverlay, vars, t)
	o, err := ApplyKustomize(ns, spinOverlay, e, t)
	assert.NotNil(t, err, fmt.Sprintf("Expected validation error. Output: %s", o))

	vars["KubernetesEnabled"] = false
	SubstituteOverlayVars(spinOverlay, vars, t)
	o, err = ApplyKustomize(ns, spinOverlay, e, t)
	assert.NotNil(t, err, fmt.Sprintf("Expected validation error. Output: %s", o))

	vars["DockerEnabled"] = false
	SubstituteOverlayVars(spinOverlay, vars, t)
	o, err = ApplyKustomize(ns, spinOverlay, e, t)
	assert.NotNil(t, err, fmt.Sprintf("Expected no errors of admission webhook. Output: %s", o))

	vars["PersistentS3Enabled"] = false
	SubstituteOverlayVars(spinOverlay, vars, t)
	o, err = ApplyKustomize(ns, spinOverlay, e, t)
	assert.Nil(t, err, fmt.Sprintf("Expected validation error. Output: %s", o))
}
