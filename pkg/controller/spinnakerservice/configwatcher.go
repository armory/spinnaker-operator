package spinnakerservice

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type configWatcher struct {
	client    client.Client
	namespace string
	queue     []v1alpha1.SpinnakerServiceInterface
}

func (c *configWatcher) MatchesConfig(meta metav1.Object) bool {
	// Get SpinnakerService in either all namespaces and single namespace
	ss, err := c.getSpinnakerServices()
	if err != nil {
		return false
	}
	added := false
	for k := range ss {
		hcm := ss[k].GetStatus().HalConfig.ConfigMap
		if hcm != nil && hcm.Name == meta.GetName() && hcm.Namespace == meta.GetNamespace() {
			c.queue = append(c.queue, ss[k])
			added = true
		}
	}
	return added
}

func (c *configWatcher) getSpinnakerServices() ([]v1alpha1.SpinnakerServiceInterface, error) {
	list := SpinnakerServiceKind.NewList()
	var opts *client.ListOptions
	if c.namespace == "" {
		opts = &client.ListOptions{}
	} else {
		opts = &client.ListOptions{Namespace: c.namespace}
	}
	err := c.client.List(context.TODO(), opts, list)
	if err != nil {
		return nil, err
	}
	return list.GetItems(), nil
}

func (c *configWatcher) Predicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			return c.MatchesConfig(e.MetaOld)
		},
	}
}

func (c *configWatcher) Map(handler.MapObject) []reconcile.Request {
	reqs := make([]reconcile.Request, 0)
	for k := range c.queue {
		reqs = append(reqs, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      c.queue[k].GetName(),
				Namespace: c.queue[k].GetNamespace(),
			},
		})
	}
	return reqs
}
