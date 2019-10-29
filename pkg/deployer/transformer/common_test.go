package transformer

import (
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/armory/spinnaker-operator/pkg/test"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"testing"
)

type testHelpers struct{}

var th = testHelpers{}

func (h *testHelpers) setupTransformer(generator Generator, spinsvcManifest string, t *testing.T, objs ...runtime.Object) (Transformer, *v1alpha2.SpinnakerService) {
	fakeClient := test.FakeClient(t, objs...)
	spinsvc := test.ManifestToSpinService(spinsvcManifest, t)
	tr, _ := generator.NewTransformer(spinsvc, fakeClient, log.Log.WithName("spinnakerservice"))
	return tr, spinsvc
}
