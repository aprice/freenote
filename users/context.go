package users

import (
	"context"
)

type contextKey int

var (
	userKey    contextKey = 0
	ownerKey   contextKey = 1
	ownerIDKey contextKey = 2
)

func FromContext(ctx context.Context) (User, bool) {
	u, ok := ctx.Value(userKey).(User)
	return u, ok
}

func NewContext(ctx context.Context, u User) context.Context {
	return context.WithValue(ctx, userKey, u)
}

func OwnerFromContext(ctx context.Context) (User, bool) {
	u, ok := ctx.Value(ownerKey).(User)
	return u, ok
}

func NewOwnerContext(ctx context.Context, u User) context.Context {
	return context.WithValue(ctx, ownerKey, u)
}
