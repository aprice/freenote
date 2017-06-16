package users

import (
	"context"
)

type contextKey int

var (
	userKey  contextKey = 1
	ownerKey contextKey = 2
)

// FromContext returns the authenticated User in the given Context, if it is set.
func FromContext(ctx context.Context) (User, bool) {
	u, ok := ctx.Value(userKey).(User)
	return u, ok
}

// NewContext returns a new Context includign the given User.
func NewContext(ctx context.Context, u User) context.Context {
	return context.WithValue(ctx, userKey, u)
}

// OwnerFromContext returns the owning User in the given Context, if it is set.
func OwnerFromContext(ctx context.Context) (User, bool) {
	u, ok := ctx.Value(ownerKey).(User)
	return u, ok
}

// NewOwnerContext returns a new Context includign the given User.
func NewOwnerContext(ctx context.Context, u User) context.Context {
	return context.WithValue(ctx, ownerKey, u)
}
