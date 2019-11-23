package secrets

import (
	"context"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestCleanup(t *testing.T) {
	ctx := NewContext(context.TODO(), nil, "ns")
	n, f, err := Decode(ctx, "encryptedFile:noop!blah")
	assert.Nil(t, err)
	assert.True(t, f)

	secCtx, ok := FromContext(ctx)
	if !assert.True(t, ok) {
		return
	}
	assert.Equal(t, 1, len(secCtx.FileCache))
	_, err = os.Stat(n)
	assert.Nil(t, err)

	Cleanup(ctx)
	_, err = os.Stat(n)
	if assert.NotNil(t, err) {
		_, ok = err.(*os.PathError)
		assert.True(t, ok)
	}
	assert.Empty(t, secCtx.FileCache)
}
