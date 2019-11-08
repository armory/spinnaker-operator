package secrets

import (
	"context"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

const encryptedSecret string = "encrypted:test"
const notASecret = "notASecret"
const badFormatSecret string = "encrypted:s3!r:us-west-2"
const decryptedValue = "mockSecret"
const didNotGetCalled = "didn't get called"

func mockDecrypt(c client.Client, namespace string, val string) (string, error) {
	return decryptedValue, nil
}

func dontCallMe(c client.Client, namespace string, val string) (string, error) {
	return didNotGetCalled, nil
}

func TestDecrypt(t *testing.T) {
	ctx := NewContext(context.TODO(), nil, "")

	// decrypt secret syntax
	v, err := decode(mockDecrypt, ctx, encryptedSecret)
	if assert.Nil(t, err) {
		assert.Contains(t, v, decryptedValue)
	}

	// don't try to decrypt a non-secret
	v, err = decode(dontCallMe, ctx, notASecret)
	if assert.Nil(t, err) {
		assert.Contains(t, v, notASecret)
		assert.NotContains(t, v, didNotGetCalled)
	}
}

func TestBadFormat(t *testing.T) {
	ctx := NewContext(context.TODO(), nil, "")

	// calling real decrypter with bad syntax should return error
	_, err := Decode(ctx, badFormatSecret)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "secret format error")
	}
}

func TestCaching(t *testing.T) {
	// cache is empty to start
	ctx := NewContext(context.TODO(), nil, "")
	c, ok := FromContext(ctx)
	if !ok {
		t.Fatalf("error getting context cache")
	}
	assert.Empty(t, c.Cache)

	// decode and store a secret
	v, err := decode(mockDecrypt, ctx, encryptedSecret)
	if assert.Nil(t, err) {
		assert.Equal(t, v, decryptedValue)
	}

	// now cache contains previous secret
	c, ok = FromContext(ctx)
	if !ok {
		t.Fatalf("error getting context cache")
	}
	assert.Contains(t, decryptedValue, c.Cache[encryptedSecret])
	assert.Equal(t, 1, len(c.Cache))

	// if we decrypt same secret again
	v, err = decode(dontCallMe, ctx, encryptedSecret)

	// decrypter method didn't get called, value returned is the old cached value
	if assert.Nil(t, err) {
		assert.Equal(t, v, decryptedValue)
		assert.NotEqual(t, v, didNotGetCalled)
	}

	// and cache still only contains one secret
	c, ok = FromContext(ctx)
	if !ok {
		t.Fatalf("error getting context cache")
	}
	assert.Contains(t, decryptedValue, c.Cache[encryptedSecret])
	assert.Equal(t, 1, len(c.Cache))
}
