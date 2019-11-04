package kubernetes

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/authorization/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type kubernetesAccountValidator struct {
	account *Account
}

func (k *kubernetesAccountValidator) Validate(spinSvc v1alpha2.SpinnakerServiceInterface, c client.Client, ctx context.Context, log logr.Logger) error {
	p, err := k.account.newKubeConfig(ctx, log)
	if err != nil {
		return err
	}
	log.Info(fmt.Sprintf("got kubeconfig file %s", p))
	return k.validateAccess(p)
}

func (k *kubernetesAccountValidator) validateAccess(kubeconfigPath string) error {
	c, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{
			CurrentContext: k.account.Auth.Context,
		}).ClientConfig()
	if err != nil {
		return err
	}
	clientset, err := kubernetes.NewForConfig(c)
	if err != nil {
		return err
	}

	// Check what the client can do
	// TODO loop through the check we need
	// Checks are by namespace if namespaces or omitNamespaces (get them before)
	// Get list of api resources (see getResources below)
	// If CRDs are listed, try them explicitly
	sars := clientset.AuthorizationV1().SelfSubjectAccessReviews()
	sar := &v1.SelfSubjectAccessReview{
		Spec: v1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &v1.ResourceAttributes{
				//Namespace:   "",
				Verb:    "*",
				Group:   "*",
				Version: "*",
				//Resource:    "",
				//Subresource: "",
				//Name:        "",
			},
			NonResourceAttributes: nil,
		},
	}
	res, err := sars.Create(sar)
	if err != nil {
		return err
	}
	if res.Status.Denied {
		return fmt.Errorf("access denied to cluster for account %s", k.account.Name)
	}
	return nil
}

//func (v *kubernetesAccountValidator) getResources(c *kubernetes.Clientset) error {
//	_, rscs, err := c.ServerGroupsAndResources()
//	if err != nil {
//		return err
//	}
//}
//
//func getAccessResourceAttributes(a Account) *v1.ResourceAttributes {
//	nss := a.Env.Namespaces
//	if len(nss) == 0 {
//
//	} else {
//
//	}
//}
//
//func getAccessResourceAttributesInNs(a Account, ns string) []*v1.ResourceAttributes {
//	if len(a.Env.Kinds) == 0 && len(a.Env.OmitKinds) == 0 {
//		return &v1.ResourceAttributes{
//			Namespace: ns,
//			Verb:      "*",
//			Group:     "*",
//			Version:   "*",
//			Resource:  "*",
//		}
//	}
//}
