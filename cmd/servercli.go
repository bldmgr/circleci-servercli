package main

import (
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

const (
	homeEnvVar    = "CIRCLE_HOME"
	hostEnvVar    = "CIRCLE_HOSTNAME"
	tokenEnvVar   = "CIRCLE_TOKEN"
	projectEnvVar = "CIRCLE_PROJECT"
)

var (
	rootCmd     *cobra.Command
	globalUsage = `Servercli is a compact and smart client that provides a simple interface that automates access to CircleCI’s API.`
	conf        *initCmd
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd = newRootCmd()
}

func newRootCmd() *cobra.Command {
	ciHome := defaultCiHome()
	conf = setConf()

	cmd := &cobra.Command{
		Use:          "servercli",
		Short:        globalUsage,
		Long:         globalUsage,
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
			return
		},
	}

	p := cmd.PersistentFlags()
	p.StringVar(&ciHome, "home", defaultCiHome(), "location of your config. Overrides $CIRCLE_HOME")

	cmd.AddCommand(
		newInitCmd(conf.host, conf.token, conf.project),
		newStatusCmd(conf.host, conf.token, conf.project),
		newTreeCmd(conf.host, conf.token, conf.project),
	)

	return cmd
}

func defaultCiHome() string {
	if home := os.Getenv(homeEnvVar); home != "" {
		return home
	}
	homeEnvPath := os.Getenv("HOME")

	return filepath.Join(homeEnvPath, ".")
}
