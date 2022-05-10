package secrets

import (
	"context"
	"fmt"
	"github.com/armory/go-yaml-tools/pkg/secrets"
	"os"
)

func init() {
	secrets.Engines["k8s"] = NewKubernetesSecretDecrypter
}

// Decode decodes a potential value into a secret
func Decode(ctx context.Context, val string) (string, bool, error) {
	if !secrets.IsEncryptedSecret(val) {
		return val, false, nil
	}

	// Get decrypter
	dec, err := secrets.NewDecrypter(ctx, val)
	if err != nil {
		return val, false, fmt.Errorf("Error creating decrypter for value '%s':\n  %w", val, err)
	}

	var v string
	c, err := FromContextWithError(ctx)
	if err != nil {
		return "", false, fmt.Errorf("Error creating secret context for value '%s':\n  %w", val, err)
	}

	// Check if in cache
	if v, ok := c.Cache[val]; ok {
		return v, false, nil
	}

	// Check if in file cache
	if v, ok := c.FileCache[val]; ok {
		return v, true, nil
	}

	v, err = dec.Decrypt()
	if err != nil {
		return "", false, fmt.Errorf("Error decrypting secret value '%s':\n  %w", val, err)
	}

	// If we could get the cache, update it
	if dec.IsFile() {
		c.FileCache[val] = v
	} else {
		c.Cache[val] = v
	}
	return v, dec.IsFile(), nil
}

// DecodeAsFile is decode with a check that the final value is a file that exists
func DecodeAsFile(ctx context.Context, val string) (string, error) {
	// We ignore the isFile return value to support old style "encrypted:" file references
	s, _, err := Decode(ctx, val)
	if err != nil {
		return "", fmt.Errorf("Error decoding string \"%s\":\n  %w", val, err)
	}
	_, err = os.Stat(s)
	if err != nil {
		return s, fmt.Errorf("Error decoding string \"%s\" into a file:\n  %w\nDid you use \"encrypted\" instead of \"encryptedFile\"?", val, err)
	}
	return s, err
}

// ShouldDecryptToValidate should we decrypt that value before sending to Halyard for validation?
// For now we decrypt everything so the operator has to be authorized to.
func ShouldDecryptToValidate(val string) bool {
	return true
}
