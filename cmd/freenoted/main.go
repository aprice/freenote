package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aprice/freenote/config"
	"github.com/aprice/freenote/rest"
	"github.com/aprice/freenote/users"
)

func main() {
	conf, err := config.Configure()
	if err != nil {
		log.Fatal(err)
	}

	err = users.InitCommonPasswords(conf)
	if err != nil {
		log.Println(err)
	}

	if conf.RecoveryMode {
		pw, err := users.RecoveryMode()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Recovery user: %q password: %q good until %s", users.RecoveryAdminName, pw, time.Now().Add(users.RecoveryPeriod))
	}

	restServer := rest.NewServer(conf)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", conf.Port), restServer))
	os.Exit(0)
}
