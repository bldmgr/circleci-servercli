package main

import (
	"fmt"
	"github.com/bldmgr/circleci"
	"github.com/bldmgr/circleci-servercli/pkg/common"
	setting "github.com/bldmgr/circleci/pkg/config"
	"github.com/jedib0t/go-pretty/v6/list"
	"github.com/spf13/cobra"
	"os"
	"strconv"
	"strings"
)

type treeCmd struct {
	host       string
	token      string
	namespace  string
	pipelineId string
	counter    int
	status     string
	action     string
	jobNumber  string
	data       string
}

const (
	treeDesc = `Single project view with more data`
	maxTree  = 10
)

func newTreeCmd(host string, token string, namespace string) *cobra.Command {
	i := &treeCmd{
		host:      host,
		token:     token,
		namespace: namespace,
	}

	cmd := &cobra.Command{
		Use:   "tree",
		Short: treeDesc,
		Long:  treeDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			return i.run()
		},
	}
	f := cmd.Flags()
	f.StringVarP(&i.pipelineId, "id", "i", "", "Pipeline Id")
	f.StringVarP(&i.status, "status", "s", "all", "Status of job")
	f.StringVarP(&i.action, "action", "a", "", "Show job steps (exp) (exp-all)")
	f.StringVarP(&i.jobNumber, "job", "j", "", "Show only job number steps")
	f.StringVarP(&i.data, "data", "d", "", "Show job output")

	return cmd
}

func displayTree(title string, content string, prefix string) {
	fmt.Println(strings.Repeat("-", len(title)+100))
	for _, line := range strings.Split(content, "\n") {
		fmt.Printf("%s%s\n", prefix, line)
	}
	fmt.Println()
}

func (cmd *treeCmd) run() error {
	jnumber, _ := strconv.Atoi(cmd.jobNumber)
	if cmd.pipelineId == "" {
		fmt.Println("No pipeline id specified")
		return nil
	}
	var returnDataSet []circleci.JobDataSteps
	l := list.NewWriter()
	lTemp := list.List{}
	lTemp.Render()

	l.SetStyle(list.StyleConnectedRounded)
	loadedConfig := setting.SetConfigYaml()

	ci, err := circleci.New(loadedConfig.Host, loadedConfig.Token, loadedConfig.Project)
	if err != nil {
		panic(err)
	}

	status := circleci.Me(ci)
	if status == false {
		fmt.Printf("Error with configuration %s -> %t \n", loadedConfig.Host, status)
		os.Exit(1)
	}
	fmt.Printf("Data is being fetched from server %s -> %t \n", loadedConfig.Host, status)
	workflows := circleci.GetPipelineWorkflows(ci, cmd.pipelineId, "none")
	project, _, _ := common.FormatProjectSlug(workflows[0].ProjectSlug)
	l.AppendItem(project)
	l.Indent()
	for w := range workflows {
		var jobs []circleci.WorkflowItem = circleci.GetWorkflowJob(ci, workflows[w].ID, "none", "i.data", "i.token")
		l.AppendItem(workflows[w].Name)
		l.Indent()
		for j := range jobs {

			if cmd.status == "all" || cmd.status == "failed" && jobs[j].Status == "failed" || cmd.status == "failed" && jobs[j].Status == "blocked" {
				returnDataSet, _, _, _ = circleci.GetConfigWithWorkflow(ci, jobs, workflows, j, w, "data")
				formatJobName := fmt.Sprintf("%d %s (%s)", jobs[j].JobNumber, jobs[j].Name, jobs[j].Status)
				l.AppendItem(formatJobName)
				//log.Printf("Config %v", returnEnvConfig[0].Sha)
				if cmd.action == "exp" && jobs[j].JobNumber == jnumber || cmd.action == "exp-all" {
					l.Indent()
					for s := range returnDataSet {
						formatStepName := fmt.Sprintf("%s %s", returnDataSet[s].ID, returnDataSet[s].Name)
						l.AppendItem(formatStepName)
						l.Indent()
						l.AppendItem(returnDataSet[s].Command)
						if cmd.data == returnDataSet[s].ID || cmd.data == "all" {
							l.Indent()
							l.AppendItem(returnDataSet[s].Output)
							l.UnIndent()
						}
						l.UnIndent()
					}
					l.UnIndent()
				}
			}
		}
		l.UnIndent()
	}

	displayTree("", l.Render(), "")
	return nil
}
