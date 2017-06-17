package config

import (
	"encoding/json"
	"os"
	"strings"
)

// Config holds all runtime configuration details.
type Config struct {
	Port               int
	TLSPort            int
	BaseURI            string
	ForceTLS           bool
	RecoveryMode       bool
	CommonPasswordList string

	LetsEncryptHosts []string
	CertFile         string
	KeyFile          string

	MailServer ConnectionInfo

	Elastic  ConnectionInfo
	Mongo    ConnectionInfo
	Postgres ConnectionInfo
	BoltDB   string
}

// NilConfig is an empty Configuration (zero value).
var NilConfig = Config{}

// ConnectionInfo encapsulates details for connecting to an outside resource.
type ConnectionInfo struct {
	Host      string
	User      string
	Password  string
	Namespace string
}

// NilConnection is an empty ConnectionInfo (zero value).
var NilConnection = ConnectionInfo{}

// Configure reads in the config file at the given path and returns it.
func Configure(path string) (Config, error) {
	c := &Config{
		Port:    80,
		TLSPort: 443,
	}
	f, err := os.Open(path)
	if err != nil {
		return NilConfig, err
	}
	err = json.NewDecoder(f).Decode(c)
	if err != nil {
		return NilConfig, err
	}
	if c.ForceTLS {
		c.BaseURI = strings.Replace(c.BaseURI, "http://", "https://", 1)
	}
	return *c, nil
}
