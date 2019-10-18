package spinnakervalidating

import (
	"github.com/stretchr/testify/assert"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"testing"
)

func TestSpinnakerServiceGVK(t *testing.T) {
	req := admission.Request{AdmissionRequest: admissionv1beta1.AdmissionRequest{
		Kind: v1.GroupVersionKind{
			Version: "v1alpha2",
			Group:   "spinnaker.io",
			Kind:    "SpinnakerService",
		},
		Resource: v1.GroupVersionResource{},
	}}
	assert.True(t, isSpinnakerRequest(req))
}
