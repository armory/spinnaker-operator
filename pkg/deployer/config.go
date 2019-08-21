package deployer

import (
	"context"
	"time"

	spinnakerv1alpha1 "github.com/armory-io/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	// "sigs.k8s.io/controller-runtime/pkg/client"
)

// IsConfigUpToDate returns true if the config in status represents the latest
// config in the service spec
func (d *Deployer) IsConfigUpToDate(instance *spinnakerv1alpha1.SpinnakerService, config runtime.Object) bool {
	hcStat := instance.Status.HalConfig
	cm, ok := config.(*corev1.ConfigMap)
	if ok {
		cmStatus := hcStat.ConfigMap
		return cmStatus != nil && cmStatus.Name == cm.ObjectMeta.Name && cmStatus.Namespace == cm.ObjectMeta.Namespace &&
			cmStatus.ResourceVersion == cm.ObjectMeta.ResourceVersion
	}
	sec, ok := config.(*corev1.Secret)
	if ok {
		secStatus := hcStat.Secret
		return secStatus != nil && secStatus.Name == sec.ObjectMeta.Name && secStatus.Namespace == sec.ObjectMeta.Namespace &&
			secStatus.ResourceVersion == sec.ObjectMeta.ResourceVersion
	}
	return false
}

func (d *Deployer) commitConfigToStatus(ctx context.Context, svc *spinnakerv1alpha1.SpinnakerService, status *spinnakerv1alpha1.SpinnakerServiceStatus, config runtime.Object) error {
	cm, ok := config.(*corev1.ConfigMap)
	if ok {
		status.HalConfig = spinnakerv1alpha1.SpinnakerFileSourceStatus{
			ConfigMap: &spinnakerv1alpha1.SpinnakerFileSourceReferenceStatus{
				Name:            cm.ObjectMeta.Name,
				Namespace:       cm.ObjectMeta.Namespace,
				ResourceVersion: cm.ObjectMeta.ResourceVersion,
			},
		}
	}
	sec, ok := config.(*corev1.Secret)
	if ok {
		status.HalConfig = spinnakerv1alpha1.SpinnakerFileSourceStatus{
			Secret: &spinnakerv1alpha1.SpinnakerFileSourceReferenceStatus{
				Name:            sec.ObjectMeta.Name,
				Namespace:       sec.ObjectMeta.Namespace,
				ResourceVersion: sec.ObjectMeta.ResourceVersion,
			},
		}
	}
	status.LastConfigurationTime = metav1.NewTime(time.Now())

	s := svc.DeepCopy()
	s.Status = *status
	// Following doesn't work (EKS) - looks like PUTting to the subresource (status) gives a 404
	// TODO Investigate issue on earlier Kubernetes version, works fine in 1.13
	return d.client.Status().Update(ctx, s)
}
