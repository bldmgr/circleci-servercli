package main

import (
	"fmt"
	"github.com/bldmgr/circleci"
	common "github.com/bldmgr/circleci-servercli/pkg/common"
	setting "github.com/bldmgr/circleci/pkg/config"
	"github.com/jedib0t/go-pretty/v6/table"
	"os"
	"strconv"
)

func main() {
	loadedConfig := setting.SetConfigYaml()

	ci, err := circleci.New(loadedConfig.Host, loadedConfig.Token, loadedConfig.Project)
	if err != nil {
		panic(err)
	}

	status := circleci.Me(ci)
	fmt.Printf("Data is being fetched from server %s -> %t \n", loadedConfig.Host, status)
	p := circleci.GetPipeline(ci, loadedConfig.Project, "web", 1)
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Project", "Number", "Pipeline Id", "Trigger Actor", "Workflow Name", "Jobs Name", "Job Number", "Job Status", "Agent Version", "Runner/VM/Image"})
	for i := range p {
		project, vcs, namespace := common.FormatProjectSlug(p[i].ProjectSlug)
		workflows := circleci.GetPipelineWorkflows(ci, p[i].ID, "none")
		for w := range workflows {
			var jobs []circleci.WorkflowItem = circleci.GetWorkflowJob(ci, workflows[w].ID, "none", "i.data", "i.token")

			for j := range jobs {
				executors := ""
				data := string(circleci.GetJobData(ci, strconv.Itoa(jobs[j].JobNumber), vcs, namespace, project, "0", ""))
				jobHost := string(data) + "\n"

				outAgent, outRunner, outVm, outImage, _, outdocker := common.ParseVariables(jobHost, "Build-agent version ", "Launch-agent version ", "Using volume:", "default", "  using image ", "Starting container ")
				if outRunner != "" {
					executors = outRunner
				}
				if outVm != "" {
					executors = outVm
				}
				if outImage != "" {
					executors = outImage
				}
				if outdocker != "" {
					executors = outdocker
				}

				t.AppendRows([]table.Row{{project, p[i].Number, p[i].ID, p[i].Trigger.Actor.Login, workflows[w].Name, jobs[j].Name, jobs[j].JobNumber, jobs[j].Status, outAgent, executors}})
			}
		}

	}
	t.Render()
}
