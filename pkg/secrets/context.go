package secrets

import (
	"context"
	"errors"
	"os"
	"sync"

	"k8s.io/client-go/rest"
)

type SecretContext struct {
	mutex      *sync.RWMutex
	Cache      map[string]string
	FileCache  map[string]string
	RestConfig *rest.Config
	Namespace  string
}

var errContextNotInitialized = errors.New("secret context not initialized")
var secretContextKey = "secretContext"

func NewContext(ctx context.Context, c *rest.Config, namespace string) context.Context {
	return context.WithValue(ctx, secretContextKey, &SecretContext{
		mutex:      &sync.RWMutex{},
		Cache:      make(map[string]string),
		FileCache:  make(map[string]string),
		RestConfig: c,
		Namespace:  namespace,
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

// Cleanup deletes any temporary file that was used
// Errors are ignored
func (s *SecretContext) Cleanup() {
	for _, f := range s.FileCache {
		os.Remove(f)
	}
	s.FileCache = make(map[string]string)
}

// Attempt to clean up secret context if it exists
func Cleanup(ctx context.Context) {
	c, ok := FromContext(ctx)
	if ok {
		c.Cleanup()
	}
}
