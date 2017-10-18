package key

import (
	"context"
	"errors"
	"sync"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-kit/kit/endpoint"
)

const (
	DefaultResolverTimeout time.Duration = 15 * time.Second
)

var (
	ErrNoKID             = errors.New("No kid in header or kid is not a string")
	ErrMismatchedKeyType = errors.New("The kid exists but is not of the given type")

	errKeyExpired = errors.New("That key has expired")
)

// ResolverRequest is the request object passed to endpoints that can fetch a key.
type ResolverRequest struct {
	Type Type
	KID  string
}

// Resolver obtains keys based on key id (kid).
type Resolver interface {
	// Key asks the Resolver to fetch a Key of a given Type.  Currently, a given
	// key id can only be of one type.  If a key id exists but is not of the given
	// type, an error is returned.
	Key(context.Context, Type, string) (Interface, error)
}

// NewResolver produces a key Resolver that uses the given endpoint to fetch key data.
func NewResolver(timeout time.Duration, endpoint endpoint.Endpoint) Resolver {
	if timeout < 1 {
		timeout = DefaultResolverTimeout
	}

	return &resolver{
		timeout:  timeout,
		endpoint: endpoint,
		cache:    make(map[string]Interface),
	}
}

// Keyfunc accepts a Resolver and produces a jwt-go Keyfunc that can
// load keys by the kid header field.  If the given token has no kid
// header field, an error is returned.
func Keyfunc(kt Type, r Resolver) jwt.Keyfunc {
	return func(t *jwt.Token) (interface{}, error) {
		if kid, ok := t.Header["kid"].(string); ok {
			key, err := r.Key(context.Background(), kt, kid)
			if err != nil {
				return nil, err
			}

			return key.Key(), nil
		}

		return nil, ErrNoKID
	}
}

type freshenResult struct {
	key Interface
	err error
}

type resolver struct {
	timeout  time.Duration
	endpoint endpoint.Endpoint

	cacheLock sync.RWMutex
	cache     map[string]Interface

	freshenLock sync.Mutex
	freshen     map[string]chan freshenResult
}

func (r *resolver) fetchKey(ctx context.Context, kt Type, kid string) (Interface, error) {
	var (
		request = &ResolverRequest{kt, kid}

		// we set the request into the context so that things like response decoders can see it
		response, err = r.endpoint(WithResolverRequest(ctx, request), request)
	)

	if err != nil {
		return nil, err
	}

	return response.(Interface), err
}

func (r *resolver) cacheKey(kid string, key Interface) {
	r.cacheLock.Lock()
	r.cache[kid] = key
	r.cacheLock.Unlock()
}

// tryCache attempts to load the key out of the cache.  This method also
// expires keys as necessary.
func (r *resolver) tryCache(kt Type, kid string) (Interface, error) {
	r.cacheLock.RLock()
	key, ok := r.cache[kid]
	r.cacheLock.RUnlock()

	if ok {
		if Expired(key, time.Now()) {
			r.cacheLock.Lock()
			delete(r.cache, kid)
			r.cacheLock.Unlock()
			return nil, errKeyExpired
		}

		if key.Type() != kt {
			return nil, ErrMismatchedKeyType
		}
	}

	return key, nil
}

func (r *resolver) freshenKey(ctx context.Context, kt Type, kid string) (Interface, error) {
	r.freshenLock.Lock()
	result, ok := r.freshen[kid]
	if !ok {
		// since this goroutine detects the missing refresh channel, it's responsible for refreshing the key
		defer func() {
			r.freshenLock.Lock()
			delete(r.freshen, kid)
			r.freshenLock.Unlock()
		}()

		result = make(chan freshenResult, 1)
		r.freshen[kid] = result

		go func() {
			key, err := r.fetchKey(ctx, kt, kid)
			if err == nil {
				r.cacheKey(kid, key)
			}

			result <- freshenResult{key, err}
		}()
	}

	r.freshenLock.Unlock()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()

	case fr := <-result:
		return fr.key, fr.err
	}
}

func (r *resolver) Key(ctx context.Context, kt Type, kid string) (Interface, error) {
	key, err := r.tryCache(kt, kid)
	if err == errKeyExpired {
		return r.freshenKey(ctx, kt, kid)
	}

	return key, err
}
