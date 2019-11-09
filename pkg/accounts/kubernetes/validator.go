package kubernetes

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/armory/spinnaker-operator/pkg/inspect"
	"github.com/armory/spinnaker-operator/pkg/secrets"
	"github.com/armory/spinnaker-operator/pkg/util"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	v1 "k8s.io/api/authorization/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	clientcmdlatest "k8s.io/client-go/tools/clientcmd/api/latest"
	clientcmdv1 "k8s.io/client-go/tools/clientcmd/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

var (
	noAuthProvidedError      = fmt.Errorf("Kubernetes auth needs to be defined")
	noKubernetesDefinedError = fmt.Errorf("Kubernetes needs to be defined")
	noValidKubeconfigError   = fmt.Errorf("no valid kubeconfig file or content found")
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
		return makeClientFromSettings(ctx, k.account.Settings)
	}
	if auth.KubeconfigFile != "" {
		return makeClientFromFile(ctx, auth.KubeconfigFile, nil)
	}
	if auth.Kubeconfig != nil {
		return makeClientFromConfigAPI(auth.Kubeconfig)
	}
	if auth.KubeconfigSecret != nil {
		return makeClientFromSecretRef(ctx, auth.KubeconfigSecret)
	}
	return nil, noAuthProvidedError
}

// makeClientFromFile loads the client config from a file path which can be a secret
func makeClientFromFile(ctx context.Context, file string, settings *authSettings) (clientcmd.ClientConfig, error) {
	file, err := secrets.DecodeAsFile(ctx, file)
	if err != nil {
		return nil, err
	}

	cfg, err := clientcmd.LoadFromFile(file)
	if err != nil {
		return nil, err
	}

	return clientcmd.NewDefaultClientConfig(*cfg, makeOverrideFromAuthSettings(cfg, settings)), nil
}

// makeClientFromSecretRef reads the client config from a Kubernetes secret in the current context's namespace
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

// makeClientFromConfigAPI makes a client config from the v1 Config (the usual format for kubeconfig) inlined
// into the CRD.
func makeClientFromConfigAPI(config *clientcmdv1.Config) (clientcmd.ClientConfig, error) {
	cfg := clientcmdapi.NewConfig()
	if err := clientcmdlatest.Scheme.Convert(config, cfg, nil); err != nil {
		return nil, nil
	}
	return clientcmd.NewDefaultClientConfig(*cfg, &clientcmd.ConfigOverrides{}), nil
}

// makeClientFromSettings makes a client config from Spinnaker settings
func makeClientFromSettings(ctx context.Context, settings map[string]interface{}) (clientcmd.ClientConfig, error) {
	aSettings := &authSettings{}
	if err := inspect.Source(aSettings, settings); err != nil {
		return nil, err
	}
	if aSettings.KubeconfigFile != "" {
		return makeClientFromFile(ctx, aSettings.KubeconfigFile, aSettings)
	}
	if aSettings.KubeconfigContents != "" {
		cfg, err := clientcmd.Load([]byte(aSettings.KubeconfigContents))
		if err != nil {
			return nil, err
		}
		return clientcmd.NewDefaultClientConfig(*cfg, makeOverrideFromAuthSettings(cfg, aSettings)), nil
	}
	return nil, noValidKubeconfigError
}

func makeOverrideFromAuthSettings(config *clientcmdapi.Config, settings *authSettings) *clientcmd.ConfigOverrides {
	overrides := &clientcmd.ConfigOverrides{}
	if settings == nil {
		return overrides
	}
	if settings.Context != "" {
		overrides.CurrentContext = settings.Context
	}
	if settings.User != "" {
		if authInfo, ok := config.AuthInfos[settings.User]; ok {
			overrides.AuthInfo = *authInfo
		}
	}
	if settings.Cluster != "" {
		if cluster, ok := config.Clusters[settings.Cluster]; ok {
			overrides.ClusterInfo = *cluster
		}
	}
	if len(settings.OAuthScopes) > 0 {
		overrides.AuthInfo = clientcmdapi.AuthInfo{
			AuthProvider: &clientcmdapi.AuthProviderConfig{
				Name: "gcp",
				Config: map[string]string{
					"scopes": strings.Join(settings.OAuthScopes, ","),
				},
			},
		}
	}
	return overrides
}

type authSettings struct {
	// User to use in the kubeconfig file
	User string `json:"user,omitempty"`
	// Context to use in the kubeconfig file if not default
	Context string `json:"context,omitempty"`
	// Cluster to use in the kubeconfig file
	Cluster        string `json:"cluster,omitempty"`
	ServiceAccount bool   `json:"serviceAccount,omitempty"`
	// Reference to a kubeconfig file
	KubeconfigFile      string   `json:"kubeconfigFile,omitempty"`
	KubeconfigContents  string   `json:"kubeconfigContents,omitempty"`
	OAuthServiceAccount string   `json:"oAuthServiceAccount,omitempty"`
	OAuthScopes         []string `json:"oAuthScopes,omitempty"`
}

func (k *kubernetesAccountValidator) validateAccess(clientConfig clientcmd.ClientConfig) error {
	cc, err := clientConfig.ClientConfig()
	if err != nil {
		return err
	}
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
