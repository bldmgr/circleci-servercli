package main

import (
	"fmt"
	"github.com/bldmgr/circleci-servercli/pkg/common"
	"github.com/spf13/viper"
	"os"
	"strings"
)

type Config struct {
	host    string `yaml:host`
	token   string `yaml:token`
	project string `yaml:project`
}

func setConf() *initCmd {
	ciHome := defaultCiHome()
	if os.Getenv(hostEnvVar) == "" {
		viper.SetConfigType("yaml")
		viper.AddConfigPath(ciHome)
		viper.SetConfigName("/ci")
		err := viper.ReadInConfig()

		if err != nil {
			fmt.Printf("%v", err)
		}

		conf := &Config{}
		err = viper.Unmarshal(conf)
		if err != nil {
			fmt.Printf("unable to decode into config struct, %v", err)
		}

		return &initCmd{
			host:    fmt.Sprintf("%v", viper.Get("circle_hostname")),
			project: fmt.Sprintf("%v", viper.Get("circle_project")),
			token:   strings.TrimSpace(common.LetsDecrypt(fmt.Sprintf("%v", viper.Get("circle_token")))),
		}
	} else {
		return &initCmd{
			host:    os.Getenv(hostEnvVar),
			token:   os.Getenv(tokenEnvVar),
			project: os.Getenv(projectEnvVar),
		}
	}
}
