package main

import (
	"fmt"
	"os"

	"github.com/aprice/freenote/cmd/freenote/commands"
)

func main() {
	err := commands.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
	// Inputs:
	// - CLI
	// - env
	// - config file
	/*
		fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		fs.StringP("user", "u", "", "Freenote user name")
		fs.StringP("password", "p", "", "Freenote password")
		fs.StringP("host", "h", "localhost:80", "Freenote server address")
		fs.Parse()
		viper.SetDefault("user", "")
		viper.SetDefault("password", "")
		viper.RegisterAlias("host", "server")
		viper.SetDefault("server", "localhost")
		viper.SetConfigName("config")
		viper.SetEnvPrefix("freenote")
		viper.AutomaticEnv()
		viper.AddConfigPath("/etc/freenote")
		viper.AddConfigPath("$HOME/.freenote")
		viper.AddConfigPath(".")
		err := viper.ReadInConfig()
		if err != nil {
			log.Fatal(err)
		}
	*/
	// Subcommands:
	// - upload
	// - download
	// - edit
	// - export
	// - backup
}
