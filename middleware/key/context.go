package key

import "context"

type resolverRequestKey struct{}

func WithResolverRequest(ctx context.Context, rr *ResolverRequest) context.Context {
	return context.WithValue(ctx, resolverRequestKey{}, rr)
}

func GetResolverRequest(ctx context.Context) *ResolverRequest {
	if rr, ok := ctx.Value(resolverRequestKey{}).(*ResolverRequest); ok {
		return rr
	}

	return nil
}
