package handler

import "context"

type userIDContextKey struct{}

func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDContextKey{}, userID)
}

func UserIDFromContext(ctx context.Context) (int64, bool) {
	v := ctx.Value(userIDContextKey{})
	userID, ok := v.(int64)
	if !ok || userID <= 0 {
		return 0, false
	}
	return userID, true
}
