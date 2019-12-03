// +build integration

package basic_mode

import (
	"encoding/json"
	"fmt"
	it "github.com/armory/spinnaker-operator/integration_tests"
	"github.com/stretchr/testify/assert"
	"os"
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

func TestKubeAccountsWithSecrets(t *testing.T) {
	// install spinnaker
	ns := "test-secrets"
	it.CreateNamespace(ns, Env, t)
	it.ApplyManifestInNs("testdata/sa.yml", ns, Env, t)
	_, gateUrl := it.DeploySpinnaker("spinnakersecrets", "testdata/spinnakerservice.yml", ns, Env, t)

	// verify the right accounts exist
	o := it.ExecuteGetRequest(fmt.Sprintf("%s/credentials", gateUrl), t)
	type c struct {
		Name string `json:"name"`
	}
	var credentials []c
	if assert.Nil(t, json.Unmarshal([]byte(o), &credentials)) {
		assert.Equal(t, "kube-sa-inline", credentials[0].Name)
	}
	it.DeleteManifestInNs("testdata/spinnakerservice.yml", ns, Env, t)
	it.DeleteNamespace(ns, Env, t)
}
