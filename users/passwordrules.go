package users

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"unicode/utf8"

	"github.com/aprice/freenote/config"
	"github.com/aprice/freenote/stringset"
)

const minPasswordLength = 8
const maxPasswordLength = 128

var commonPasswords = stringset.New()

// ErrPasswordTooShort indicates a password that failed validation for minimum length.
var ErrPasswordTooShort = fmt.Errorf("password does not meet minimum length %d", minPasswordLength)

// ErrPasswordTooLong indicates a password that failed validation for maximum length.
var ErrPasswordTooLong = fmt.Errorf("password exceeds maximum length %d", maxPasswordLength)

// ErrPasswordTooRepetitive indicates a password that failed validation for character repetition.
var ErrPasswordTooRepetitive = errors.New("password is too repetitive")

// ErrPasswordTooCommon indicates a password that failed validation for common passwords.
var ErrPasswordTooCommon = errors.New("password is too common")

var initOnce = new(sync.Once)

// InitCommonPasswords reads in the configured list of common passwords.
func InitCommonPasswords(conf config.Config) error {
	var err error
	initOnce.Do(func() {
		if conf.CommonPasswordList != "" {
			var f *os.File
			f, err = os.Open(conf.CommonPasswordList)
			if err != nil {
				return
			}
			err = readCommonPasswords(f)
		}
	})
	return err
}

func readCommonPasswords(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		pw := scanner.Text()
		if ValidatePassword(pw) == nil {
			commonPasswords.Add(pw)
		}
	}
	return scanner.Err()
}

// ValidatePassword against the password rules.
func ValidatePassword(password string) error {
	byteLen := len(password)
	charLen := utf8.RuneCountInString(password)
	if byteLen < minPasswordLength {
		return ErrPasswordTooShort
	}
	if byteLen > maxPasswordLength {
		return ErrPasswordTooLong
	}
	chars := stringset.New()
	for _, ch := range password {
		chars.Add(string(ch))
	}
	if chars.Count() <= (charLen / 2) {
		return ErrPasswordTooRepetitive
	}
	if commonPasswords == nil {
		return nil
	}
	if commonPasswords.Contains(password) {
		return ErrPasswordTooCommon
	}
	return nil
}
