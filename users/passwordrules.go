package users

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"unicode/utf8"

	"github.com/aprice/freenote/config"
	"github.com/aprice/freenote/stringset"
)

const minPasswordLength = 8
const maxPasswordLength = 128

var commonPasswords = stringset.New()

var ErrPasswordTooShort = fmt.Errorf("password does not meet minimum length %d", minPasswordLength)
var ErrPasswordTooLong = fmt.Errorf("password exceeds maximum length %d", maxPasswordLength)
var ErrPasswordTooRepetitive = errors.New("password is too repetitive")
var ErrPasswordTooCommon = errors.New("password is too common")

func InitCommonPasswords(conf config.Config) error {
	if conf.CommonPasswordList != "" {
		f, err := os.Open(conf.CommonPasswordList)
		if err != nil {
			return err
		}
		return ReadCommonPasswords(f)
	}
	return nil
}

func ReadCommonPasswords(r io.Reader) error {
	if commonPasswords != nil {
		return nil
	}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		pw := scanner.Text()
		if ValidatePassword(pw) == nil {
			commonPasswords.Add(pw)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

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
