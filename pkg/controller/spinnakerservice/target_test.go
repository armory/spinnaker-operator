package spinnakerservice

import (
	"testing"
	corev1 "k8s.io/api/core/v1"
	spinnakerv1alpha1 "github.com/armory-io/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	kruntime "k8s.io/apimachinery/pkg/runtime"
	appsv1 "k8s.io/api/apps/v1beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/armory-io/spinnaker-operator/pkg/generated"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestSetTarget(t *testing.T) {
	var gen = &generated.SpinnakerGeneratedConfig{
		Config: map[string]generated.ServiceConfig{
			"orca": generated.ServiceConfig{
				Deployment: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "ns1",
					},
				},
				Service: &corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "ns1",
					},
				},
				Resources: []runtime.Object{
					&corev1.ConfigMap{Data: map[string]string{"abc": "def"}},
				},
			},
		},
	}
	s := kruntime.NewScheme()
	svc := spinnakerv1alpha1.SpinnakerService{
		ObjectMeta: metav1.ObjectMeta{ Namespace: "ns2" },
	}
	tg := &targetTransformer{ svc: svc }
	err := tg.TransformManifests(s, nil, gen, nil)
	if assert.Nil(t, err) {
		assert.Equal(t, "ns2", gen.Config["orca"].Deployment.ObjectMeta.Namespace)
		assert.Equal(t, "ns2", gen.Config["orca"].Service.ObjectMeta.Namespace)
		r, ok := gen.Config["orca"].Resources[0].(metav1.Object)
		if assert.True(t, ok) {
			assert.Equal(t, "ns2", r.GetNamespace())
		}
	}
}
