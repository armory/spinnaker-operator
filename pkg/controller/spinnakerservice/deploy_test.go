package spinnakerservice

import (
	"testing"

	spinnakerv1alpha1 "github.com/armory-io/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	cmp "github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func TestStatusCheck(t *testing.T) {
	h := spinnakerv1alpha1.SpinnakerFileSource{
		ConfigMap: &spinnakerv1alpha1.SpinnakerFileSourceReference{Name: "test"},
	}
	g := h.DeepCopy()
	instance := &spinnakerv1alpha1.SpinnakerService{
		Spec:   spinnakerv1alpha1.SpinnakerServiceSpec{HalConfig: h},
		Status: spinnakerv1alpha1.SpinnakerServiceStatus{HalConfig: *g},
	}
	assert.True(t, cmp.Equal(instance.Status.HalConfig, instance.Spec.HalConfig))
}
