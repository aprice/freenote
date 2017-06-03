package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Port               int
	TLSPort            int
	BaseURI            string
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

var NilConfig = Config{}

type ConnectionInfo struct {
	Host      string
	User      string
	Password  string
	Namespace string
}

var NilConnection = ConnectionInfo{}

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
	return *c, err
}
