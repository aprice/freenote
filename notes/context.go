package notes

import (
	"context"

	uuid "github.com/satori/go.uuid"
)

type contextKey int

var (
	noteKey   contextKey = 1
	noteIDKey contextKey = 2
)

// FromContext reads a Note from the given Context, if it is set.
func FromContext(ctx context.Context) (Note, bool) {
	n, ok := ctx.Value(noteKey).(Note)
	return n, ok
}

// NewContext creates a new Context including the given Note and returns it.
func NewContext(ctx context.Context, u Note) context.Context {
	return context.WithValue(ctx, noteKey, u)
}

// IDFromContext reads a note ID from the given Context, if it is set.
func IDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(noteIDKey).(uuid.UUID)
	return id, ok
}

// NewIDContext creates a new Context including the given note ID and returns it.
func NewIDContext(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, noteIDKey, id)
}
