package main

import (
	"fmt"
	common "github.com/bldmgr/circleci-servercli/pkg/common"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
)

type initCmd struct {
	host    string
	token   string
	project string
	data    string
}

const (
	initDesc = `Sets up local environment to work with circle`
)

func newInitCmd(host string, token string, project string) *cobra.Command {
	ciHome := defaultCiHome()
	i := &initCmd{
		host:    host,
		token:   token,
		project: project,
	}

	cmd := &cobra.Command{
		Use:   "init",
		Short: initDesc,
		Long:  initDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			return i.run(ciHome)
		},
	}
	return cmd
}

func (i *initCmd) run(ciHome string) error {
	fmt.Println("")
	h, p := common.SetInput()
	t := common.GetInput("Token: ")
	c := &Config{
		host:    h,
		token:   t,
		project: p,
	}

	viper.Set(hostEnvVar, c.host)
	viper.Set(tokenEnvVar, c.token)
	viper.Set(projectEnvVar, c.project)
	viper.SetConfigType("yaml")
	err := viper.WriteConfigAs(ciHome + "/ci.yaml")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Circle has been configured at %s and storing token securely \n", ciHome)
	return nil
}
