package kubernetes

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidate(t *testing.T) {
	k := &AccountType{}
	a := k.newAccount()
	a.Auth.KubeconfigFile = "encrypted:nop!v:sssss"
	v := a.NewValidator()
	err := v.Validate(nil, nil, context.TODO(), nil)
	assert.Nil(t, err)
}
