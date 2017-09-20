package key

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-kit/kit/endpoint"
)

// TransientKey is a key which can expire.  The return from endpoints used by
// Resolver can implement this interface.
type TransientKey interface {
	Key() interface{}
	Expiry() time.Time
}

// Resolver obtains keys based on key id (kid).
type Resolver interface {
	Key(context.Context, string) (interface{}, error)
}

func Wrap(endpoint endpoint.Endpoint) Resolver {
	return &resolver{
		cache: make(map[string]interface{}),
	}
}

type resolver struct {
	endpoint endpoint.Endpoint

	cacheLock sync.RWMutex
	cache     map[string]interface{}
}

func (r *resolver) tryCache(kid string) (interface{}, bool) {
	r.cacheLock.RLock()
	key, ok := r.cache[kid]
	r.cacheLock.RUnlock()
	return key, ok
}

func (r *resolver) putKey(kid string, key interface{}) {
	r.cacheLock.Lock()
	r.cache[kid] = key
	r.cacheLock.Unlock()
}

func (r *resolver) freshenKey(ctx context.Context, kid string) (interface{}, error) {
	key, err := r.endpoint(ctx, kid)
	if err != nil {
		return nil, err
	}

	if transientKey, ok := key.(TransientKey); ok {
		key = transientKey.Key()

		duration := transientKey.Expiry().Sub(time.Now())
		if duration < 1 {
			return nil, fmt.Errorf("Key %s has already expired", kid)
		}

		r.putKey(kid, key)
		time.AfterFunc(duration, func() {
			// TODO: Handle errors here, possibly just logging
			r.freshenKey(context.Background(), kid)
		})
	} else {
		r.putKey(kid, key)
	}

	return key, nil
}

func (r *resolver) Key(ctx context.Context, kid string) (interface{}, error) {
	key, ok := r.tryCache(kid)
	if ok {
		return key, nil
	}

	return r.freshenKey(ctx, kid)
}
