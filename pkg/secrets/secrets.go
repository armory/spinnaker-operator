package secrets

import (
	"context"
	"github.com/armory/go-yaml-tools/pkg/secrets"
	"io/ioutil"
	"os"
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

type decrypter func(val string) (string, error)

func decode(decrypt decrypter, ctx context.Context, val string) (string, error) {
	if !isSecretEncrypted(val) {
		return val, nil
	}

	var v string
	c, cacheOk := FromContext(ctx)
	if cacheOk {
		// Check if in cache
		if v, ok := (*c)[val]; ok {
			return v, nil
		}
	}
	v, err := decrypt(val)
	if err != nil {
		return "", err
	}
	if cacheOk {
		// If we could get the cache, update it
		(*c)[val] = v
	}
	return v, nil
}

func isSecretEncrypted(str string) bool {
	return strings.HasPrefix(str, "encrypted:")
}

func decryptFunc(val string) (string, error) {
	secretDecrypter := secrets.NewDecrypter(val)
	return secretDecrypter.Decrypt()
}
