package kubernetes

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Secret struct {
	name string
	key  string
}

type SecretDecrypter struct {
	params    map[string]string
	client    client.Client
	namespace string
}

func NewSecretDecrypter(c client.Client, ns string, params map[string]string) *SecretDecrypter {
	return &SecretDecrypter{params: params, client: c, namespace: ns}
}

func (s *SecretDecrypter) Decrypt() (string, error) {
	sec, err := ParseSecret(s.params)
	if err != nil {
		return "", err
	}
	return s.fetchSecret(sec)
}

func ParseSecret(params map[string]string) (Secret, error) {
	var sec Secret

	name, ok := params["n"]
	if !ok {
		return Secret{}, fmt.Errorf("secret format error - 'n' for name is required")
	}
	sec.name = name

	key, ok := params["k"]
	if !ok {
		return Secret{}, fmt.Errorf("secret format error - 'k' for secret key is required")
	}
	sec.key = key
	return sec, nil
}

func (sd *SecretDecrypter) fetchSecret(secret Secret) (string, error) {
	sec := &v1.Secret{}
	if err := sd.client.Get(context.TODO(), client.ObjectKey{Name: secret.name, Namespace: sd.namespace}, sec); err != nil {
		return "", err
	}
	if d, ok := sec.Data[secret.key]; ok {
		return string(d), nil
	}
	return "", fmt.Errorf("cannot find key %s in secret %s", secret.key, secret.name)
}
