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
	v13 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	if err := k.validateSettings(ctx, log); err != nil {
		return err
	}
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
	// Get namespaces. The test is analogous to what is done in Halyard
	// We want to keep it short so any improvement should remain short (e.g. not a request per namespace)
	_, err = clientset.CoreV1().Namespaces().List(v13.ListOptions{})
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("unable to verify access to account %s", k.account.Name))
	}
	return nil
}

func (k *kubernetesAccountValidator) validateSettings(ctx context.Context, log logr.Logger) error {
	nss, err := inspect.GetStringArray(k.account.Settings, "namespaces")
	if err != nil {
		nss = make([]string, 0)
	}
	omitNss, err := inspect.GetStringArray(k.account.Settings, "omitNamespaces")
	if err != nil {
		omitNss = make([]string, 0)
	}
	if len(nss) > 0 && len(omitNss) > 0 {
		return fmt.Errorf("At most one of \"namespaces\" and \"omitNamespaces\" can be supplied.")
	}
	return nil
}
