package spinnakervalidating

import (
	"github.com/armory/spinnaker-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	admissionv1 "k8s.io/api/admission/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"testing"
)

func init() {
	TypesFactory = test.TypesFactory
}

func TestSpinnakerServiceGVK(t *testing.T) {
	req := admission.Request{AdmissionRequest: admissionv1.AdmissionRequest{
		Kind: v1.GroupVersionKind{
			Version: "v1alpha2",
			Group:   "spinnaker.io",
			Kind:    "SpinnakerService",
		},
		Resource: v1.GroupVersionResource{},
	}}
	assert.True(t, isSpinnakerRequest(req))
}
