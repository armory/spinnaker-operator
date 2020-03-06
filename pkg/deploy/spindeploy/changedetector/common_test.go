package changedetector

import (
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"testing"
)

type testHelpers struct {
	TypesFactory interfaces.TypesFactory
}

var th = testHelpers{
	TypesFactory: test.TypesFactory,
}

func (th *testHelpers) setupChangeDetector(generator Generator, t *testing.T, objs ...runtime.Object) ChangeDetector {
	fakeClient := test.FakeClient(t, objs...)
	ch, err := generator.NewChangeDetector(fakeClient, log.Log.WithName("spinnakerservice"))
	assert.Nil(t, err)
	return ch
}
