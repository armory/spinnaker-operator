package transformer

import (
	"testing"

	"github.com/armory/spinnaker-operator/pkg/api/interfaces"
	"github.com/armory/spinnaker-operator/pkg/api/test"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type TestHelpers struct {
	TypesFactory interfaces.TypesFactory
}

var th = TestHelpers{
	TypesFactory: interfaces.DefaultTypesFactory,
}

func (h *TestHelpers) SetupTransformerFromSpinFile(generator Generator, spinsvcManifest string, t *testing.T, objs ...runtime.Object) (Transformer, interfaces.SpinnakerService) {
	fakeClient := test.FakeClient(t, objs...)
	spinsvc := test.ManifestFileToSpinService(spinsvcManifest, t)
	tr, _ := generator.NewTransformer(spinsvc, fakeClient, log.Log.WithName("spinnakerservice"), runtime.NewScheme())
	return tr, spinsvc
}

func (h *TestHelpers) SetupTransformerFromSpinText(generator Generator, spinText string, t *testing.T, objs ...runtime.Object) (Transformer, interfaces.SpinnakerService) {
	fakeClient := test.FakeClient(t, objs...)
	spinsvc := test.ManifestToSpinService(spinText, t)
	tr, _ := generator.NewTransformer(spinsvc, fakeClient, log.Log.WithName("spinnakerservice"), runtime.NewScheme())
	return tr, spinsvc
}
