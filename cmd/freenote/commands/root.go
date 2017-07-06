package commands

import (
	"fmt"
	"path"

	"github.com/aprice/freenote/client"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "freenote",
	Short: "Freenote is a CLI tool for interacting with the Freenote server",
	Long: `
A simple CLI tool to make working with Freenote from the command line or other
applications easier. Add a macro to any editor to send your files to Freenote!
©2017 Adrian Price. http://github.com/aprice/freenote`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.HelpFunc()(cmd, args)
	},
}

var cfgFile string
var username string
var password string
var host string

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

// nolint: gas
func init() {
	initViper()

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.freenote/config.json)")
	rootCmd.PersistentFlags().StringVarP(&username, "user", "u", "", "Freenote account name")
	rootCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "Freenote account password")
	rootCmd.PersistentFlags().StringVarP(&host, "host", "H", "", "Freenote server address")
	viper.BindPFlag("user", rootCmd.PersistentFlags().Lookup("user"))
	viper.BindPFlag("password", rootCmd.PersistentFlags().Lookup("password"))
	viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host"))
	viper.SetDefault("user", "")
	viper.SetDefault("password", "")
	viper.RegisterAlias("host", "server")
	viper.SetDefault("server", "localhost")
}

func initClient() (*client.Client, error) {
	return client.New(viper.Get("user").(string),
		viper.Get("password").(string),
		viper.Get("host").(string))
}

func initViper() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath("/etc/freenote")

		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
		} else {
			viper.AddConfigPath(path.Join(home, ".freenote"))
		}

		viper.AddConfigPath(".")
		viper.SetConfigName("config")
	}

	viper.SetEnvPrefix("freenote")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		//fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		fmt.Println("Error reading config file: ", err)
	}
}
