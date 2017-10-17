package key

import (
	"context"
	"errors"
	"sync"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-kit/kit/endpoint"
)

var (
	ErrNoKID = errors.New("No kid in header or kid is not a string")
)

// Resolver obtains keys based on key id (kid).
type Resolver interface {
	Key(context.Context, string) (Interface, error)
}

// NewResolver produces a key Resolver that uses the given endpoint to fetch key data.
func NewResolver(endpoint endpoint.Endpoint) Resolver {
	return &resolver{
		endpoint: endpoint,
		cache:    make(map[string]Interface),
	}
}

// Keyfunc accepts a Resolver and produces a jwt-go Keyfunc that can
// load keys by the kid header field.  If the given token has no kid
// header field, an error is returned.
func Keyfunc(r Resolver) jwt.Keyfunc {
	return func(t *jwt.Token) (interface{}, error) {
		if kid, ok := t.Header["kid"].(string); ok {
			key, err := r.Key(context.Background(), kid)
			if err != nil {
				return nil, err
			}

			return key.Key(), nil
		}

		return nil, ErrNoKID
	}
}

type resolver struct {
	endpoint endpoint.Endpoint

	cacheLock sync.RWMutex
	cache     map[string]Interface

	freshenLock sync.Mutex
	freshen     map[string]context.Context
}

func (r *resolver) fetchKey(ctx context.Context, kid string) (Interface, error) {
	response, err := r.endpoint(ctx, kid)
	if err != nil {
		return nil, err
	}

	return response.(Interface), err
}

func (r *resolver) tryCache(kid string) (Interface, bool) {
	r.cacheLock.RLock()
	key, ok := r.cache[kid]
	r.cacheLock.RUnlock()
	return key, ok
}

func (r *resolver) putKey(key Interface) {
	r.cacheLock.Lock()
	r.cache[key.KID()] = key
	r.cacheLock.Unlock()
}

func (r *resolver) freshenKey(ctx context.Context, kid string) (Interface, error) {
	r.freshenLock.Lock()
	ctx, ok := r.freshen[kid]
	r.freshenLock.Unlock()

	if ok {
		// another goroutine is currently working on freshening the key
		// the context will always be cancelable
		<-ctx.Done()
	} else {
	}
}

func (r *resolver) Key(ctx context.Context, kid string) (Interface, error) {
	key, ok := r.tryCache(kid)
	if ok {
		return key, nil
	}

	return r.freshenKey(ctx, kid)
}
