package secrets

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBadFormat(t *testing.T) {
	ctx := NewContext(context.TODO(), nil, "")
	defer Cleanup(ctx)

	// calling real decrypter with bad syntax should return error
	_, _, err := Decode(ctx, "encrypted:s3!r:us-west-2")
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "secret format error")
	}
}

func TestCaching(t *testing.T) {
	// cache is empty to start
	ctx := NewContext(context.TODO(), nil, "")
	defer Cleanup(ctx)

	c, ok := FromContext(ctx)
	if !ok {
		t.Fatalf("error getting context cache")
	}
	assert.Empty(t, c.Cache)

	v, _, err := Decode(ctx, "encrypted:noop!myvalue")
	assert.Nil(t, err)
	assert.Equal(t, "myvalue", v)
	assert.NotEmpty(t, c.Cache)
	assert.Contains(t, "myvalue", c.Cache["encrypted:noop!myvalue"])
}
