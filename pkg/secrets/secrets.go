package secrets

import (
	"context"
	"fmt"
	"github.com/armory/go-yaml-tools/pkg/secrets"
	"github.com/armory/spinnaker-operator/pkg/secrets/kubernetes"
	"io/ioutil"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

// Decode decodes a potential value into a secret
func Decode(ctx context.Context, val string) (string, error) {
	return decode(decryptFunc, ctx, val)
}

func DecodeAsFile(ctx context.Context, val string) (string, error) {
	if !isSecretEncrypted(val) {
		// Check the file exists
		_, err := os.Stat(val)
		return val, err
	}
	c, err := decode(decryptFunc, ctx, val)
	if err != nil {
		return "", err
	}
	f, err := ioutil.TempFile("", "operator-")
	if err != nil {
		return "", err
	}
	defer f.Close()

	f.Write([]byte(c))
	return f.Name(), nil
}

type decrypter func(c client.Client, namespace string, val string) (string, error)

func decode(decrypt decrypter, ctx context.Context, val string) (string, error) {
	if !isSecretEncrypted(val) {
		return val, nil
	}

	var v string
	c, err := FromContextWithError(ctx)
	if err != nil {
		return "", err
	}
	// Check if in cache
	if v, ok := c.Cache[val]; ok {
		return v, nil
	}
	v, err = decrypt(c.Client, c.Namespace, val)
	if err != nil {
		return "", err
	}
	// If we could get the cache, update it
	c.Cache[val] = v
	return v, nil
}

func isSecretEncrypted(str string) bool {
	return strings.HasPrefix(str, "encrypted:")
}

func decryptFunc(c client.Client, namespace string, val string) (string, error) {
	secretDecrypter, err := NewDecrypter(c, namespace, val)
	if err != nil {
		return "", err
	}
	return secretDecrypter.Decrypt()
}

func NewDecrypter(c client.Client, namespace string, encryptedSecret string) (secrets.Decrypter, error) {
	engine, params := secrets.ParseTokens(encryptedSecret)
	switch engine {
	case "s3":
		return secrets.NewS3Decrypter(params), nil
	case "vault":
		return secrets.NewVaultDecrypter(params), nil
	case "k8s":
		return kubernetes.NewSecretDecrypter(c, namespace, params), nil
	case "noop":
		params["v"] = encryptedSecret[len("encrypted:noop!v:"):]
		return NewNop(params), nil
	default:
		return nil, fmt.Errorf("secret engine %s not found", engine)
	}
}
