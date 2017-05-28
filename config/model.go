package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Port               int
	BaseURI            string
	RecoveryMode       bool
	CommonPasswordList string

	Elastic ConnectionInfo
	Mongo   ConnectionInfo
	BoltDB  string
}

var NilConfig = Config{}

type ConnectionInfo struct {
	Host      string
	User      string
	Password  string
	Namespace string
}

var NilConnection = ConnectionInfo{}

func Configure() (Config, error) {
	c := new(Config)
	f, err := os.Open("config.json")
	if err != nil {
		return NilConfig, err
	}
	err = json.NewDecoder(f).Decode(c)
	return *c, err
}
