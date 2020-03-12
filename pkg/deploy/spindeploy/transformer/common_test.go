package transformer

import (
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/test"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"testing"
)

type testHelpers struct {
	TypesFactory interfaces.TypesFactory
}

var th = testHelpers{
	TypesFactory: interfaces.DefaultTypesFactory,
}

func (h *testHelpers) setupTransformerFromSpinFile(generator Generator, spinsvcManifest string, t *testing.T, objs ...runtime.Object) (Transformer, interfaces.SpinnakerService) {
	fakeClient := test.FakeClient(t, objs...)
	spinsvc := test.ManifestFileToSpinService(spinsvcManifest, t)
	tr, _ := generator.NewTransformer(spinsvc, fakeClient, log.Log.WithName("spinnakerservice"))
	return tr, spinsvc
}

func (h *testHelpers) setupTransformerFromSpinText(generator Generator, spinText string, t *testing.T, objs ...runtime.Object) (Transformer, interfaces.SpinnakerService) {
	fakeClient := test.FakeClient(t, objs...)
	spinsvc := test.ManifestToSpinService(spinText, t)
	tr, _ := generator.NewTransformer(spinsvc, fakeClient, log.Log.WithName("spinnakerservice"))
	return tr, spinsvc
}
