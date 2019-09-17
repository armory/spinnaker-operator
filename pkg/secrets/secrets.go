package secrets

import (
	"context"
	"github.com/armory/go-yaml-tools/pkg/secrets"
	"strings"
)

// Decode decodes a potential value into a secret
func Decode(ctx context.Context, val string) (string, error) {
	return decode(decryptFunc, ctx, val)
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
