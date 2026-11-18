package sim

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
)

// Live environment variables. The default `go test` suite never contacts
// TEST; set GORO_LIVE_TEST=1 and GORO_LIVE_ACCOUNT_FILE to opt in.
const (
	EnvLiveTest        = "GORO_LIVE_TEST"
	EnvLiveAccountFile = "GORO_LIVE_ACCOUNT_FILE"
)

var (
	ErrLiveDisabled  = errors.New("live tests disabled: set GORO_LIVE_TEST=1")
	ErrLiveNoAccount = errors.New("live tests need GORO_LIVE_ACCOUNT_FILE pointing at a kgap pool JSON")
	ErrLiveNotKgap   = errors.New("live tests may only use kgap pool characters")
)

var kgapUser = regexp.MustCompile(`^kgap[0-9]+$`)

// Account is the uncommitted kgap JSON (username, password, login_address).
// Password is never logged.
type Account struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	LoginAddress string `json:"login_address"`
}

// CheckLiveEnv reports whether a live TEST run is allowed. It does not
// print credentials or host addresses.
func CheckLiveEnv() error {
	if os.Getenv(EnvLiveTest) != "1" {
		return ErrLiveDisabled
	}
	path := os.Getenv(EnvLiveAccountFile)
	if path == "" {
		return ErrLiveNoAccount
	}
	acct, err := LoadAccount(path)
	if err != nil {
		return err
	}
	return acct.RequireKgap()
}

// LoadAccount reads a kgap JSON file. Callers must not log Password.
func LoadAccount(path string) (Account, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Account{}, err
	}
	var acct Account
	if err := json.Unmarshal(data, &acct); err != nil {
		return Account{}, fmt.Errorf("live account file: %w", err)
	}
	return acct, nil
}

// RequireKgap rejects any username that is not a kgap pool character,
// including the live account name the rules forbid.
func (a Account) RequireKgap() error {
	if !kgapUser.MatchString(a.Username) {
		return ErrLiveNotKgap
	}
	return nil
}

// IsKgapUsername reports whether name is a kgap pool login.
func IsKgapUsername(name string) bool {
	return kgapUser.MatchString(name)
}
