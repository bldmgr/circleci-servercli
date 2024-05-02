package common

import (
	"bufio"
	"path/filepath"
	"strings"
)

func FormatProjectSlug(projectSlug string) (project string, vcs string, namespace string) {
	// Split vcs/namespace/project
	out, project := filepath.Split(projectSlug)
	s := strings.TrimRight(out, "/")
	x, namespace := filepath.Split(s)
	vcs = strings.TrimRight(x, "/")

	return project, vcs, namespace
}

func removeText(data string, start string, end string, number int) (out string) {
	format := strings.Replace(data, start, "", number)
	out = strings.Replace(format, end, "", number)
	return out
}

func ParseVariables(data string, agent string, runner string, volume string, vm string, image string, docker string) (outAgent string, outRunner string, outVm string, outImage string, outVolume string, outDocker string) {
	theReader := strings.NewReader(data)
	scanner := bufio.NewScanner(theReader)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, agent) == true {
			outAgent = strings.Replace(line, agent, "", -1)
		}
		if strings.Contains(line, runner) == true {
			outRunner = strings.Replace(line, runner, "", -1)
		}
		if strings.Contains(line, vm) == true {
			outVm = removeText(line, "VM '", "' has been created", -1)
		}
		if strings.Contains(line, volume) == true {
			outVolume = strings.Replace(line, volume, "", -1)
		}
		if strings.Contains(line, image) == true {
			outImage = strings.Replace(line, image, "", -1)
		}
		if strings.Contains(line, docker) == true {
			outDocker = strings.Replace(line, docker, "", -1)
		}
	}

	return outAgent, outRunner, outVm, outImage, outVolume, outDocker
}
