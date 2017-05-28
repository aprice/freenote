package users

import (
	"crypto/rand"
	"encoding/base64"
	"time"
)

const RecoveryPeriod = 30 * time.Minute

var recoveryMode = struct {
	user    User
	expires time.Time
}{}

const RecoveryAdminName = "_admin"

// RecoveryMode initializes recovery mode, creating a temporary admin account,
// and returning the password to the account.
func RecoveryMode() (string, error) {
	passRaw := make([]byte, 12)
	_, err := rand.Read(passRaw)
	if err != nil {
		return "", err
	}

	passString := base64.StdEncoding.EncodeToString(passRaw)
	password, err := NewPassword(passString)
	if err != nil {
		return "", err
	}

	recoveryMode.user = User{
		Username:    RecoveryAdminName,
		DisplayName: "Recovery Admin",
		Access:      LevelRecovery,
		Password:    password,
	}
	recoveryMode.expires = time.Now().Add(RecoveryPeriod)
	return passString, nil
}

// AuthenticateAdmin authenticates the given password against the temporary admin
// account, if any.
func AuthenticateAdmin(password string) (User, error) {
	if recoveryMode.expires.Before(time.Now()) {
		return User{}, ErrAuthenticationFailed
	}
	ok, err := recoveryMode.user.Password.Verify(password)
	if err != nil {
		return User{}, err
	}
	if ok {
		return recoveryMode.user, nil
	}
	return User{}, ErrAuthenticationFailed
}
