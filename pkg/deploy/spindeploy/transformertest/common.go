package transformertest

import (
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/deploy/spindeploy/transformer"
	"github.com/armory/spinnaker-operator/pkg/test"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
)

func SetupTransformerFromSpinFile(generator transformer.Generator, spinsvcManifest string, t *testing.T, objs ...runtime.Object) (transformer.Transformer, interfaces.SpinnakerService) {
	fakeClient := test.FakeClient(t, objs...)
	spinsvc := test.ManifestFileToSpinService(spinsvcManifest, t)
	tr, _ := generator.NewTransformer(spinsvc, fakeClient, log.Log.WithName("spinnakerservice"), runtime.NewScheme())
	return tr, spinsvc
}

func SetupTransformerFromSpinText(generator transformer.Generator, spinText string, t *testing.T, objs ...runtime.Object) (transformer.Transformer, interfaces.SpinnakerService) {
	fakeClient := test.FakeClient(t, objs...)
	spinsvc := test.ManifestToSpinService(spinText, t)
	tr, _ := generator.NewTransformer(spinsvc, fakeClient, log.Log.WithName("spinnakerservice"), runtime.NewScheme())
	return tr, spinsvc
}
