package secrets

import (
	"context"
	"errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SecretContext struct {
	Cache     map[string]string
	Client    client.Client
	Namespace string
}

var errContextNotInitialized = errors.New("secret context not initialized")
var secretContextKey = "secretContext"

func NewContext(ctx context.Context, c client.Client, namespace string) context.Context {
	return context.WithValue(ctx, secretContextKey, &SecretContext{
		Cache:     make(map[string]string),
		Client:    c,
		Namespace: namespace,
	})
}

func FromContext(ctx context.Context) (*SecretContext, bool) {
	c, ok := ctx.Value(secretContextKey).(*SecretContext)
	return c, ok
}

func FromContextWithError(ctx context.Context) (*SecretContext, error) {
	if c, ok := FromContext(ctx); ok {
		return c, nil
	}
	return nil, errContextNotInitialized
}
