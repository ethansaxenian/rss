package contextkeys

import "context"

type contextKey string

const ContextKeyPath = contextKey("path")

func GetRoutePathCtx(ctx context.Context) string {
	path, ok := ctx.Value(ContextKeyPath).(string)
	if !ok {
		return ""
	}

	return path
}

func WithRoutePathCtx(ctx context.Context, path string) context.Context {
	return context.WithValue(ctx, ContextKeyPath, path)
}
