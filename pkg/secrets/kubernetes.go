package secrets

import (
	"context"
	"fmt"
	"strings"

	"github.com/armory/go-yaml-tools/pkg/secrets"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

type KubernetesDecrypter struct {
	name       string
	key        string
	restConfig *rest.Config
	namespace  string
	isFile     bool
}

func NewKubernetesSecretDecrypter(ctx context.Context, isFile bool, params string) (secrets.Decrypter, error) {
	c, err := FromContextWithError(ctx)
	if err != nil {
		return nil, err
	}
	k := &KubernetesDecrypter{restConfig: c.RestConfig, namespace: c.Namespace, isFile: isFile}
	if err := k.parse(params); err != nil {
		return nil, err
	}
	return k, nil
}

func (k *KubernetesDecrypter) Decrypt() (string, error) {
	client, err := corev1.NewForConfig(k.restConfig)
	if err != nil {
		return "", fmt.Errorf("Error creating kubernetes client:\n  %w", err)
	}
	sec, err := client.Secrets(k.namespace).Get(context.TODO(), k.name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("Error reading secret with name '%s' from kubernetes:\n  %w", k.name, err)
	}
	if d, ok := sec.Data[k.key]; ok {
		if k.isFile {
			return secrets.ToTempFile(d)
		}
		return string(d), nil
	}
	return "", fmt.Errorf("Cannot find key %s in secret %s", k.key, k.name)
}

func (s *KubernetesDecrypter) IsFile() bool {
	return s.isFile
}

func (k *KubernetesDecrypter) parse(params string) error {
	name, key, err := ParseKubernetesSecretParams(params)
	k.name = name
	k.key = key
	return err
}

func ParseKubernetesSecretParams(params string) (string, string, error) {
	var name, key string
	tokens := strings.Split(params, "!")
	for _, element := range tokens {
		kv := strings.Split(element, ":")
		if len(kv) != 2 {
			return "", "", fmt.Errorf("Secret format error - 'n' for name is required, 'k' for secret key is required - got '%s'", element)
		}

		switch kv[0] {
		case "n":
			name = kv[1]
		case "k":
			key = kv[1]
		}
	}

	if name == "" {
		return "", "", fmt.Errorf("Secret format error - 'n' for name is required")
	}
	if key == "" {
		return "", "", fmt.Errorf("Secret format error - 'k' for secret key is required")
	}
	return name, key, nil
}
