package transformer

import (
	spinnakerv1alpha1 "github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/armory/spinnaker-operator/pkg/halconfig"
	"github.com/armory/spinnaker-operator/pkg/test"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"testing"
)

type testHelpers struct{}

var th = testHelpers{}

func (h *testHelpers) setupTransformer(generator Generator, t *testing.T) (Transformer, *spinnakerv1alpha1.SpinnakerService, *halconfig.SpinnakerConfig) {
	fakeClient := fake.NewFakeClient()
	return h.setupTransformerWithFakeClient(generator, fakeClient, t)
}

func (h *testHelpers) setupTransformerWithFakeClient(generator Generator, fakeClient client.Client, t *testing.T) (Transformer, *spinnakerv1alpha1.SpinnakerService, *halconfig.SpinnakerConfig) {
	spinSvc, hc, _ := test.SetupSpinnakerService("testdata/spinsvc.json", "testdata/halconfig.yml", t)
	test.AddProfileToConfig("gate", "testdata/profile_gate.yml", hc, t)
	tr, _ := generator.NewTransformer(spinSvc, hc, fakeClient, log.Log.WithName("spinnakerservice"))
	return tr, spinSvc, hc
}
