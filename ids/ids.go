package ids

import (
	"encoding/base64"

	"github.com/satori/go.uuid"
)

// UUID format: 32 chars
// B64 UUID: 22 chars

// ToBase64 returns the given UUID as a base64 (URL format) encoded string.
func ToBase64(id uuid.UUID) string {
	return base64.RawURLEncoding.EncodeToString(id.Bytes())
}

// Base64ToUUID attempts to convert a base64 encoded string into a UUID.
func Base64ToUUID(b64 string) (uuid.UUID, error) {
	bytes, err := base64.RawURLEncoding.DecodeString(b64)
	if err == nil {
		return uuid.FromBytes(bytes)
	}
	bytes, err = base64.RawStdEncoding.DecodeString(b64)
	if err == nil {
		return uuid.FromBytes(bytes)
	}
	return uuid.Nil, err
}

// ParseID attempts to convert the given string to a UUID. The input can be
// either a standard UUID text representation, or base64 (URL format) encoded
// bytes.
func ParseID(in string) (uuid.UUID, error) {
	id, err := uuid.FromString(in)
	if err == nil {
		return id, nil
	}
	return Base64ToUUID(in)
}
