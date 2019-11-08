package kubernetes

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/armory/spinnaker-operator/pkg/secrets"
	"github.com/armory/spinnaker-operator/pkg/util"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	v1 "k8s.io/api/authorization/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	noAuthProvidedError      = fmt.Errorf("Kubernetes auth needs to be defined")
	noKubernetesDefinedError = fmt.Errorf("Kubernetes needs to be defined")
)

type kubernetesAccountValidator struct {
	account *Account
}

func (k *kubernetesAccountValidator) Validate(spinSvc v1alpha2.SpinnakerServiceInterface, c client.Client, ctx context.Context, log logr.Logger) error {
	config, err := k.makeClient(ctx)
	if err != nil {
		return err
	}
	return k.validateAccess(config)
}

func (k *kubernetesAccountValidator) makeClient(ctx context.Context) (clientcmd.ClientConfig, error) {
	auth := k.account.Auth
	if auth == nil {
		// Attempt from settings
		return nil, noKubernetesDefinedError
	}

	if auth.KubeconfigFile != "" {
		return makeClientFromFile(ctx, auth.KubeconfigFile)
	}
	if auth.Kubeconfig != nil {
		//auth.Kubeconfig
		//return clientcmd.NewDefaultClientConfig(*auth.Kubeconfig, nil), nil
	}
	if auth.KubeconfigSecret != nil {
		return makeClientFromSecretRef(ctx, auth.KubeconfigSecret)
	}
	return nil, noAuthProvidedError
}

func makeClientFromFile(ctx context.Context, file string) (clientcmd.ClientConfig, error) {
	file, err := secrets.DecodeAsFile(ctx, file)
	if err != nil {
		return nil, err
	}

	cfg, err := clientcmd.LoadFromFile(file)
	if err != nil {
		return nil, err
	}
	return clientcmd.NewDefaultClientConfig(*cfg, nil), nil
}

func makeClientFromSecretRef(ctx context.Context, ref *v1alpha2.SecretInNamespaceReference) (clientcmd.ClientConfig, error) {
	sc, err := secrets.FromContextWithError(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to make kubeconfig file")
	}
	str, err := util.GetSecretContent(sc.Client, sc.Namespace, ref.Name, ref.Key)
	if err != nil {
		return nil, err
	}
	return clientcmd.NewClientConfigFromBytes([]byte(str))
}

func (k *kubernetesAccountValidator) validateAccess(clientConfig clientcmd.ClientConfig) error {
	cc, err := clientConfig.ClientConfig()
	if err != nil {
		return err
	}
	//clientConfig.
	//c, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
	//	&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
	//	&clientcmd.ConfigOverrides{
	//		CurrentContext: k.account.Auth.Context,
	//	}).ClientConfig()
	//if err != nil {
	//	return err
	//}
	clientset, err := kubernetes.NewForConfig(cc)
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
