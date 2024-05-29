package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/bldmgr/circleci"
	"github.com/bldmgr/circleci-servercli/pkg/common"
	setting "github.com/bldmgr/circleci/pkg/config"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"os"
	"strconv"
)

type dashboardCmd struct {
	host     string
	token    string
	project  string
	counter  int
	page     int
	status   string
	actor    string
	expand   string
	queryaws string
}

const (
	dashboardDesc = `Checking if connection is successful`
	maxdashboard  = 10
)

func newStatusCmd(host string, token string, project string) *cobra.Command {
	i := &dashboardCmd{
		host:    host,
		token:   token,
		project: project,
	}

	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: dashboardDesc,
		Long:  dashboardDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			return i.run()
		},
	}
	f := cmd.Flags()
	f.StringVarP(&i.project, "project", "p", "all", "Project Name")
	f.StringVarP(&i.status, "status", "s", "all", "Status of job")
	f.StringVarP(&i.actor, "actor", "a", "all", "Actor name")
	f.StringVarP(&i.expand, "expand", "e", "",
		"Expand job")
	f.StringVarP(&i.queryaws, "query-aws", "q", "", "Query AWS")

	return cmd
}

func GetRunningInstances(client *ec2.EC2) (*ec2.DescribeInstancesOutput, error) {
	result, err := client.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("instance-state-name"),
				Values: []*string{
					aws.String("running"),
				},
			},
		},
	})

	if err != nil {
		return nil, err
	}

	return result, err
}

type InstanceItem struct {
	InstanceId      string `json:"instance_id"`
	InstanceName    string `json:"instance_name"`
	PublicDnsName   string `json:"public_dns_name"`
	ImageId         string `json:"image_id"`
	InstanceType    string `json:"instance_type"`
	LaunchTime      string `json:"launch_time"`
	Architecture    string `json:"architecture"`
	PlatformDetails string `json:"platform_details"`
	VpcId           string `json:"vpc_id"`
	SubnetId        string `json:"subnet_id"`
}

type TestMetadata struct {
	Classname string  `json:"classname"`
	File      string  `json:"file"`
	Name      string  `json:"name"`
	Result    string  `json:"result"`
	Message   string  `json:"message"`
	RunTime   float64 `json:"run_time"`
	Source    string  `json:"source"`
}

type TestDataJson struct {
	Id                string `json:"id"`
	Totals            int    `json:"totals"`
	Failure           int    `json:"failure"`
	FailurePercentage string `json:"fpercentage"`
	Success           int    `json:"success"`
	SuccessPercentage string `json:"spercentage"`
	Error             int    `json:"error"`
	ErrorPercentage   string `json:"epercentage"`
	Skipped           int    `json:"skipped"`
	SkippedPercentage string `json:"kpercentage"`
}

func checkForValue(value string, instanceItem []InstanceItem) string {
	for q := range instanceItem {
		if value == instanceItem[q].InstanceName {
			return instanceItem[q].ImageId
		}
	}
	return ""
}

func (cmd *dashboardCmd) run() error {
	instanceName := ""
	instanceItem := make([]InstanceItem, 0)
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
	cmd.counter = 0
	if cmd.project == "all" {
		cmd.page = 1
	} else {
		cmd.page = 20
	}

	if cmd.queryaws == "aws" {
		sess, err := session.NewSession(&aws.Config{
			Region: aws.String("us-east-1")},
		)

		// Create EC2 service client
		svc := ec2.New(sess)

		runningInstances, err := GetRunningInstances(svc)
		if err != nil {
			fmt.Printf("Couldn't retrieve running instances: %v", err)
		}

		for _, reservation := range runningInstances.Reservations {
			for _, instance := range reservation.Instances {

				for i := 0; i < len(instance.Tags); i++ {
					if *instance.Tags[i].Key == "Name" {
						instanceName = *instance.Tags[i].Value
					}
				}
				instanceItem = append(instanceItem, InstanceItem{
					InstanceId:      *instance.InstanceId,
					InstanceName:    instanceName,
					ImageId:         *instance.ImageId,
					InstanceType:    *instance.InstanceType,
					Architecture:    *instance.Architecture,
					PlatformDetails: *instance.PlatformDetails,
					VpcId:           *instance.VpcId,
					SubnetId:        *instance.SubnetId,
					PublicDnsName:   *instance.PublicDnsName,
				})

			}
		}

	}

	p := circleci.GetPipeline(ci, loadedConfig.Project, "web", cmd.page)
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	if cmd.expand == "exp" {
		t.AppendHeader(table.Row{"Project", "Number", "Pipeline Id", "Branch", "Trigger Actor", "Workflow Name", "Jobs Name", "Job Number", "Job Status", "Agent Version", "Runner/VM/Image", "AMI ID"})
	} else {
		t.AppendHeader(table.Row{"Project", "Number", "Pipeline Id", "Branch", "Trigger Actor", "Workflow Name", "Jobs Name", "Job Number", "Job Status", "Started At", "Stopped At", "Tests"})
	}

	//GetTestMetadata(ci CI, job_id string, vsc string, namespace string, project string, output string, page int)

	for i := range p {
		project, vcs, namespace := common.FormatProjectSlug(p[i].ProjectSlug)
		if project == cmd.project && cmd.counter != maxdashboard || cmd.project == "all" {
			cmd.counter++
			if cmd.actor == "all" || cmd.actor == p[i].Trigger.Actor.Login {
				workflows := circleci.GetPipelineWorkflows(ci, p[i].ID, "none")
				for w := range workflows {
					var jobs []circleci.WorkflowItem = circleci.GetWorkflowJob(ci, workflows[w].ID, "none", "i.data", "i.token")
					for j := range jobs {
						counter := 0
						cnt_failure := 0
						cnt_success := 0
						cnt_error := 0
						cnt_skipped := 0
						items := make([]TestMetadata, 0)
						TestMessage := ""
						payload := circleci.GetTestMetadata(ci, strconv.Itoa(jobs[j].JobNumber), vcs, namespace, project, "", 1)
						if len(payload) != 0 {
							for i := range payload {
								counter++
								if payload[i].Result == "failure" {
									cnt_failure++
									items = append(items, TestMetadata{
										Classname: payload[i].Classname,
										File:      payload[i].File,
										Name:      payload[i].Name,
										Result:    payload[i].Result,
										Message:   payload[i].Message,
										RunTime:   payload[i].RunTime,
										Source:    payload[i].Source,
									})
								}
								if payload[i].Result == "success" {
									cnt_success++
								}
								if payload[i].Result == "error" {
									cnt_error++
									items = append(items, TestMetadata{
										Classname: payload[i].Classname,
										File:      payload[i].File,
										Name:      payload[i].Name,
										Result:    payload[i].Result,
										Message:   payload[i].Message,
										RunTime:   payload[i].RunTime,
										Source:    payload[i].Source,
									})
								}
								if payload[i].Result == "skipped" {
									cnt_skipped++
								}
							}
							if cnt_failure != 0 {
								TestMessage = fmt.Sprintf("%d test failed out of %d", cnt_failure, counter)
							} else {
								TestMessage = fmt.Sprintf("%d tests are passing", counter)
							}
						}

						if cmd.status == "all" || cmd.status == "failed" && jobs[j].Status == "failed" || cmd.status == "failed" && jobs[j].Status == "blocked" || cmd.status == "running" && jobs[j].Status == "running" {
							if cmd.expand == "exp" {
								data := string(circleci.GetJobData(ci, strconv.Itoa(jobs[j].JobNumber), vcs, namespace, project, "0", ""))
								jobHost := string(data) + "\n"

								outAgent, executors := common.ParseVariables(jobHost, "Build-agent version ", "Launch-agent version ", "Using volume:", "default", "  using image ", "Starting container ")

								if cmd.queryaws == "aws" && cmd.status == "running" {
									amazonImageId := checkForValue(executors, instanceItem)
									t.AppendRows([]table.Row{{project, p[i].Number, p[i].ID, p[i].Vcs.Branch, p[i].Trigger.Actor.Login, workflows[w].Name, jobs[j].Name, jobs[j].JobNumber, jobs[j].Status, outAgent, executors, amazonImageId}})
								} else {
									t.AppendRows([]table.Row{{project, p[i].Number, p[i].ID, p[i].Vcs.Branch, p[i].Trigger.Actor.Login, workflows[w].Name, jobs[j].Name, jobs[j].JobNumber, jobs[j].Status, outAgent, executors}})
								}
							} else {
								t.AppendRows([]table.Row{{project, p[i].Number, p[i].ID, p[i].Vcs.Branch, p[i].Trigger.Actor.Login, workflows[w].Name, jobs[j].Name, jobs[j].JobNumber, jobs[j].Status, jobs[j].StartedAt, jobs[j].StoppedAt, TestMessage}})
							}
						}
					}
				}
				t.AppendSeparator()
			}
		}

	}
	t.SetStyle(table.StyleColoredGreenWhiteOnBlack)
	t.Render()
	return nil
}
