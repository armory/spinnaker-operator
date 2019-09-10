package secrets

import "context"

// Decode decodes a potential value into a secret

func Decode(ctx context.Context, val string) (string, error) {
	if !isSecretEncrypted(val) {
		return val, nil
	}

	var v string
	c, ok := FromContext(ctx)
	if ok {
		// Check if in cache
		if v, ok = (*c)[val]; ok {
			return v, nil
		}
	}
	v, err := decrypt(val)
	if err != nil {
		return "", err
	}
	if ok {
		// If we could get the cache, update it
		(*c)[val] = v
	}
	return v, nil
}

func isSecretEncrypted(str string) bool {
	return false
}

func decrypt(val string) (string, error) {
	// TODO
	return val, nil
}
