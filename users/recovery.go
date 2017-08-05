package users

import (
	"time"
)

// RecoveryPeriod is the Duration that the recovery user will be able to log in
// after the server starts up.
const RecoveryPeriod = 30 * time.Minute

var recoveryMode = struct {
	user    User
	expires time.Time
}{}

// RecoveryAdminName is the username for the recovery user.
const RecoveryAdminName = "_admin"

// RecoveryMode initializes recovery mode, creating a temporary admin account,
// and returning the password to the account.
func RecoveryMode() (string, error) {
	passString, password, err := RandomPassword(16)
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

// RecoveryUser returns the recovery user, if any, and true if it is active or
// false otherwise.
func RecoveryUser() (User, bool) {
	if recoveryMode.expires.Before(time.Now()) {
		return User{}, false
	}
	return recoveryMode.user, true
}
