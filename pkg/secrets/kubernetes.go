package secrets

import (
	"context"
	"fmt"
	"github.com/armory/go-yaml-tools/pkg/secrets"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

type KubernetesDecrypter struct {
	name      string
	key       string
	client    client.Client
	namespace string
	isFile    bool
}

func NewSecretDecrypter(ctx context.Context, isFile bool, params string) (secrets.Decrypter, error) {
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
	sec := &v1.Secret{}
	if err := k.client.Get(context.TODO(), client.ObjectKey{Name: k.name, Namespace: k.namespace}, sec); err != nil {
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
	tokens := strings.Split(params, "!")
	for _, element := range tokens {
		kv := strings.Split(element, ":")
		if len(kv) == 2 {
			switch kv[0] {
			case "n":
				k.name = kv[1]
			case "k":
				k.key = kv[1]
			}
		}
	}

	if k.name == "" {
		return fmt.Errorf("secret format error - 'n' for name is required")
	}
	if k.key == "" {
		return fmt.Errorf("secret format error - 'k' for secret key is required")
	}
	return nil
}
