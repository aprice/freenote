package notes

import (
	"context"

	uuid "github.com/satori/go.uuid"
)

type contextKey int

var (
	noteKey   contextKey = 0
	noteIDKey contextKey = 1
)

func FromContext(ctx context.Context) (Note, bool) {
	n, ok := ctx.Value(noteKey).(Note)
	return n, ok
}

func NewContext(ctx context.Context, u Note) context.Context {
	return context.WithValue(ctx, noteKey, u)
}

func IDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(noteIDKey).(uuid.UUID)
	return id, ok
}

func NewIDContext(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, noteIDKey, id)
}
