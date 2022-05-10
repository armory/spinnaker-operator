package changedetectortest

import (
	"testing"

	"github.com/armory/spinnaker-operator/pkg/api/deploy/spindeploy/changedetector"
	"github.com/armory/spinnaker-operator/pkg/api/test"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func SetupChangeDetector(generator changedetector.DetectorGenerator, t *testing.T, objs ...runtime.Object) changedetector.ChangeDetector {
	fakeClient := test.FakeClient(t, objs...)
	ch, err := generator.NewChangeDetector(fakeClient, log.Log.WithName("spinnakerservice"), &record.FakeRecorder{}, runtime.NewScheme())
	assert.Nil(t, err)
	return ch
}
