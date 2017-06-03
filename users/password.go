package users

import (
	"bytes"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/pbkdf2"
)

const currentPasswordVersion uint8 = 1

// Password encapsulates all of the data necessary to hash
// and validate passwords securely.
type Password struct {
	Version uint8
	Hash    []byte
	Salt    []byte
}

// NewPassword creates a new salt and hash for the given password, using the
// current latest password version.
func NewPassword(password string) (*Password, error) {
	scheme := schemes[currentPasswordVersion]
	salt, err := scheme.Salt()
	if err != nil {
		return nil, err
	}
	pwd := &Password{
		Version: currentPasswordVersion,
		Hash:    scheme.Hash([]byte(password), salt),
		Salt:    salt,
	}
	return pwd, nil
}

// RandomPassword generates a random password of the given length using a CPRNG.
func RandomPassword(length int) (string, *Password, error) {
	// DecodedLen returns the max byte length, we want to ensure we get at least
	// length characters, so we add 1 to be safe.
	passRaw := make([]byte, base64.RawStdEncoding.DecodedLen(length)+1)
	_, err := rand.Read(passRaw)
	if err != nil {
		return "", nil, err
	}
	passString := base64.RawStdEncoding.EncodeToString(passRaw)
	passString = passString[:length]
	pw, err := NewPassword(passString)
	return passString, pw, err
}

// Verify that the given password string matches this hash.
func (p Password) Verify(input string) (bool, error) {
	// First make sure the password itself is valid and not corrupted during store/load
	if p.Version == 0 || int(p.Version) >= len(schemes) {
		return false, fmt.Errorf("unknown password version: %d", p.Version)
	}
	si := schemes[p.Version].GetInfo()
	if len(p.Salt) != si.SaltLen || len(p.Hash) != si.HashLen {
		return false, fmt.Errorf("invalid password, bad salt or hash length")
	}
	// Then validate the input matches
	provided := schemes[p.Version].Hash([]byte(input), p.Salt)
	return bytes.Equal(provided, p.Hash), nil
}

// NeedsUpdate returns true if the password scheme is out of date
// and needs updating. Note that this cannot be done automatically,
// because we can't get the plaintext password from the old hash to
// generate a new one.
func (p Password) NeedsUpdate() bool {
	return p.Version < currentPasswordVersion
}

// Password schemes, internal use only

// passwordScheme defines a secure password storage mechanism.
type passwordScheme interface {
	// Salt generates the user's unique password salt
	Salt() ([]byte, error)
	// Hash generates a secure hash from a password and salt
	Hash(password, salt []byte) []byte
	// GetInfo returns metadata about this password scheme
	GetInfo() schemeInfo
}

// schemeInfo provides details for sanity checking password data.
type schemeInfo struct {
	Version uint8
	SaltLen int
	HashLen int
}

// Any scheme used should be registered here.
var schemes = []passwordScheme{
	nil,
	(*passwordV1)(nil),
}

// Password Version 1
// - PBKDF2-HMAC-SHA512
// - 16 byte salt
// - 5,000 iterations
// - 32 byte hash
// - Based on NIST and OWASP recommendations for 2016 and drafts for 2017
type passwordV1 struct{}

func (p *passwordV1) Salt() ([]byte, error) {
	salt := make([]byte, p.GetInfo().SaltLen)
	_, err := rand.Read(salt)
	return salt, err
}

func (p *passwordV1) Hash(password, salt []byte) []byte {
	return pbkdf2.Key(password, salt, 5000, p.GetInfo().HashLen, sha512.New)
}

func (p *passwordV1) GetInfo() schemeInfo {
	return schemeInfo{Version: 1, SaltLen: 16, HashLen: 32}
}
