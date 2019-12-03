// +build integration

package basic_mode

import (
	it "github.com/armory/spinnaker-operator/integration_tests"
	"os"
	"testing"
)

var Env *it.TestEnv
var TestNs string

func init() {
	TestNs = "test-spinnaker-basic-mode"
}

// TestMain is the entry point for all tests
func TestMain(m *testing.M) {
	println("============================ Tests preparation start =========================")
	e, err := it.SetupEnv("../../deploy/crds", "../../deploy/operator/basic", TestNs)
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

func TestInstallSpinnakerBasicMode(t *testing.T) {
	it.DeploySpinnaker("spinnaker", "testdata/spinnakerservice.yml", TestNs, Env, t)
	it.DeleteManifestInNs("testdata/spinnakerservice.yml", TestNs, Env, t)
}
