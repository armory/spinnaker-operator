package secrets

import "context"

type SecretCache = map[string]string

var cacheKey = "secretCache"

func NewContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, cacheKey, &SecretCache{})
}

func FromContext(ctx context.Context) (*SecretCache, bool) {
	c, ok := ctx.Value(cacheKey).(*SecretCache)
	return c, ok
}
