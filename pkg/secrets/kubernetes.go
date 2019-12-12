package secrets

import (
	"context"
	"fmt"
	"github.com/armory/go-yaml-tools/pkg/secrets"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strings"
)

type KubernetesDecrypter struct {
	name      string
	key       string
	client    *kubernetes.Clientset
	namespace string
	isFile    bool
}

func NewKubernetesSecretDecrypter(ctx context.Context, isFile bool, params string) (secrets.Decrypter, error) {
	c, err := FromContextWithError(ctx)
	if err != nil {
		return nil, err
	}
	k := &KubernetesDecrypter{client: c.Client, namespace: c.Namespace, isFile: isFile}
	if err := k.parse(params); err != nil {
		return nil, err
	}
	return k, nil
}

func (k *KubernetesDecrypter) Decrypt() (string, error) {
	sec, err := k.client.CoreV1().Secrets(k.namespace).Get(k.name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	if d, ok := sec.Data[k.key]; ok {
		if k.isFile {
			return secrets.ToTempFile(d)
		}
		return string(d), nil
	}
	return "", fmt.Errorf("cannot find key %s in secret %s", k.key, k.name)
}

func (s *KubernetesDecrypter) IsFile() bool {
	return s.isFile
}

func (k *KubernetesDecrypter) parse(params string) error {
	name, key := ParseKubernetesSecretParams(params)
	k.name = name
	k.key = key
	if k.name == "" {
		return fmt.Errorf("secret format error - 'n' for name is required")
	}
	if k.key == "" {
		return fmt.Errorf("secret format error - 'k' for secret key is required")
	}
	return nil
}

func ParseKubernetesSecretParams(params string) (string, string) {
	var name, key string
	tokens := strings.Split(params, "!")
	for _, element := range tokens {
		kv := strings.Split(element, ":")
		if len(kv) == 2 {
			switch kv[0] {
			case "n":
				name = kv[1]
			case "k":
				key = kv[1]
			}
		}
	}
	return name, key
}
