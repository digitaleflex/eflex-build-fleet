package controlplane

import "context"

type nodeContextKey struct{}

func withNodeID(r interface{ Context() context.Context }, id string) context.Context {
	return context.WithValue(r.Context(), nodeContextKey{}, id)
}

func nodeIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(nodeContextKey{}).(string)
	return id
}
