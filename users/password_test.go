package users

import (
	"crypto/sha256"
	"crypto/sha512"
	"hash"
	"testing"

	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/sha3"
)

func TestNewPassword(t *testing.T) {
	pw, err := NewPassword("swordfish")
	if err != nil {
		t.Error(err)
	}

	ok, err := pw.Verify("swordfish")
	if err != nil {
		t.Error(err)
	} else if !ok {
		t.Error("valid password not accepted")
	}

	ok, err = pw.Verify("catfish")
	if err != nil {
		t.Error(err)
	} else if ok {
		t.Error("invalid password accepted")
	}
}

func TestRandomPassword(t *testing.T) {
	s, pw, err := RandomPassword(11)
	if err != nil {
		t.Error(err)
	}
	if len(s) != 11 {
		t.Error("random password not requested length")
	}

	ok, err := pw.Verify(s)
	if err != nil {
		t.Error(err)
	} else if !ok {
		t.Error("valid password not accepted")
	}

	ok, err = pw.Verify("catfish")
	if err != nil {
		t.Error(err)
	} else if ok {
		t.Error("invalid password accepted")
	}
}

func TestV1(t *testing.T) {
	v1 := (*passwordV1)(nil)
	salt, err := v1.Salt()
	if err != nil {
		t.Error(err)
	}
	if len(salt) != v1.GetInfo().SaltLen {
		t.Error("incorrect salt length")
	}
	h := v1.Hash([]byte("swordfish"), salt)
	if len(h) != v1.GetInfo().HashLen {
		t.Error("incorrect hash length")
	}
}

func BenchmarkV1(b *testing.B) {
	var h []byte
	v1 := (*passwordV1)(nil)
	salt, err := v1.Salt()
	if err != nil {
		b.Error(err)
	}
	pass := []byte("swordfish")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h = v1.Hash(pass, salt)
	}
	if len(h) < 32 {
		b.Error("incorrect output hash length")
	}
}

func BenchmarkHashAlgos(b *testing.B) {
	// Benchmark hashing algorithms to measure computation cost.
	// Target a hashing algo with ~10ms cost on modern hardware.
	testCases := []struct {
		name string
		h    func() hash.Hash
		iter int
	}{
		{"sha-256-5ki", sha256.New, 5000},
		{"sha-256-10ki", sha256.New, 10000},
		{"sha-256-15ki", sha256.New, 15000},
		{"sha-256-20ki", sha256.New, 20000},
		{"sha-512-5ki", sha512.New, 5000},
		{"sha-512-10ki", sha512.New, 10000},
		{"sha-512-15ki", sha512.New, 15000},
		{"sha-512-20ki", sha512.New, 20000},
		{"sha3-256-5ki", sha3.New256, 5000},
		{"sha3-256-10ki", sha3.New256, 10000},
		{"sha3-256-15ki", sha3.New256, 15000},
		{"sha3-256-20ki", sha3.New256, 20000},
		{"sha3-512-5ki", sha3.New512, 5000},
		{"sha3-512-10ki", sha512.New, 10000},
		{"sha3-512-15ki", sha512.New, 15000},
		{"sha3-512-20ki", sha512.New, 20000},
	}
	for _, tc := range testCases {
		password := []byte("swordfish")
		salt := []byte("NaClNaClNaClNaCl")
		b.Run(tc.name, func(bb *testing.B) {
			var h []byte
			for i := 0; i < bb.N; i++ {
				h = pbkdf2.Key(password, salt, tc.iter, 32, tc.h)
			}
			if len(h) < 32 {
				bb.Error("incorrect output hash length")
			}
		})
	}
}
