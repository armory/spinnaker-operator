// +build integration

package basic_mode

import (
	"encoding/json"
	"fmt"
	it "github.com/armory/spinnaker-operator/integration_tests"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

var Env *it.TestEnv

// TestMain is the entry point for all tests
func TestMain(m *testing.M) {
	println("============================ Tests preparation start =========================")
	e, err := it.SetupEnv("../../deploy/crds", "../../deploy/operator/cluster", "spinnaker-operator")
	if err != nil {
		println(err)
		os.Exit(1)
	}
	Env = e
	println("============================ Tests preparation end ===========================")
	code := m.Run()
	println("============================ Tests cleanup start =========================")
	Env.Cleanup()
	println("============================ Tests cleanup end ===========================")
	os.Exit(code)
}

func TestInstallUpgradeUninstall(t *testing.T) {
	// install spinnaker
	ns := "test-spinnaker-cluster-mode"
	spinName := "spinnaker"
	it.CreateNamespace(ns, Env, t)
	defer it.DeleteNamespace(ns, Env, t)
	it.ApplyManifestInNs("testdata/sa.yml", ns, Env, t)
	_, gateUrl := it.DeploySpinnaker(spinName, "testdata/spinnakerservice.yml", ns, Env, t)
	if t.Failed() {
		return
	}

	// verify the right accounts exist by querying gate credentials endpoint
	o := it.ExecuteGetRequest(fmt.Sprintf("%s/credentials", gateUrl), t)
	type c struct {
		Name string `json:"name"`
	}
	var credentials []c
	if assert.Nil(t, json.Unmarshal([]byte(o), &credentials)) {
		assert.Equal(t, "kube-sa-inline", credentials[0].Name)
	}

	// upgrade
	v, err := it.RunCommand(fmt.Sprintf("%s -n %s get spinsvc %s -o=jsonpath='{.status.version}'", Env.KubectlPrefix(), ns, spinName))
	if assert.Nil(t, err) {
		assert.Equal(t, "1.17.0", strings.TrimSpace(v))
	}
	it.DeploySpinnaker(spinName, "testdata/spinnakerservice_upgrade.yml", ns, Env, t)
	if t.Failed() {
		return
	}
	v, err = it.RunCommand(fmt.Sprintf("%s -n %s get spinsvc %s -o=jsonpath='{.status.version}'", Env.KubectlPrefix(), ns, spinName))
	if assert.Nil(t, err) {
		assert.Equal(t, "1.17.1", strings.TrimSpace(v))
	}

	// uninstall
	o, err = it.RunCommand(fmt.Sprintf("%s -n %s delete spinsvc %s", Env.KubectlPrefix(), ns, spinName))
	assert.Nil(t, err, o)
}
