package deployer

import (
	"testing"

	spinnakerv1alpha1 "github.com/armory-io/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestConfigUpToDate(t *testing.T) {
	d := Deployer{}
	cm := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "myconfig",
			Namespace:       "ns1",
			ResourceVersion: "123456",
		},
	}
	h := spinnakerv1alpha1.SpinnakerFileSourceStatus{
		ConfigMap: &spinnakerv1alpha1.SpinnakerFileSourceReferenceStatus{
			Name:      "myconfig",
			Namespace: "ns1",
		},
	}
	instance := &spinnakerv1alpha1.SpinnakerService{
		Status: spinnakerv1alpha1.SpinnakerServiceStatus{HalConfig: h},
	}
	assert.False(t, d.IsConfigUpToDate(instance, &cm))

	h.ConfigMap.ResourceVersion = "123456"
	assert.True(t, d.IsConfigUpToDate(instance, &cm))
}
