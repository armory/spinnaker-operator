package deployer

import (
	"testing"

	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestPatch(t *testing.T) {
	d := Deployer{}
	s1 := corev1.Service{
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Port:       7000,
					Protocol:   "TCP",
					TargetPort: intstr.FromInt(7001),
				},
			},
		},
	}
	s2 := corev1.Service{
		Spec: corev1.ServiceSpec{
			ClusterIP: "10.100.175.108",
		},
	}
	err := d.patch(&s1, &s2)
	if assert.Nil(t, err) {
		assert.Equal(t, 2, len(s2.Spec.Ports))
		assert.Equal(t, "10.100.175.108", s2.Spec.ClusterIP)
	}
}

// func TestUpdate(t *testing.T) {
// 	d := Deployer{}
// 	n := &corev1.Service{
// 		Spec: corev1.ServiceSpec{
// 			Type: corev1.ServiceTypeLoadBalancer,
// 		},
// 	}
// 	e := &corev1.Service{
// 		ObjectMeta: metav1.ObjectMeta{

// 		},
// 		Spec: corev1.ServiceSpec{

// 		},
// 	}

// 	assert.NotNil(t, d)
// 	d.updateObject(n, e)
// 	assert.Equal(t, e.Spec.Type, corev1.ServiceTypeLoadBalancer)
// }
