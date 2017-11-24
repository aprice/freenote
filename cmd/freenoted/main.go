package main

import (
	"log"
	"os"
	"os/signal"
	"time"

	flag "github.com/spf13/pflag"

	"github.com/aprice/freenote/config"
	"github.com/aprice/freenote/server"
	"github.com/aprice/freenote/stats"
	"github.com/aprice/freenote/users"
)

func main() {
	stats.Run()

	cfile := flag.StringP("config", "c", "/etc/freenoted/config.json", "Config file path")
	recovery := flag.Bool("recovery", false, "Admin recovery mode")
	flag.Parse()

	conf, err := config.Configure(*cfile)
	if err != nil {
		log.Fatal(err)
	}

	conf.RecoveryMode = conf.RecoveryMode || *recovery

	err = users.InitCommonPasswords(conf)
	if err != nil {
		log.Println(err)
	}

	if conf.RecoveryMode {
		var pw string
		pw, err = users.RecoveryMode()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Recovery user: %q password: %q good until %s", users.RecoveryAdminName, pw, time.Now().Add(users.RecoveryPeriod))
	}

	restServer, err := server.New(conf)
	if err != nil {
		log.Fatal(err)
	}
	restServer.Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	restServer.Stop()
}
