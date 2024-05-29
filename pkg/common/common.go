package common

import (
	"bufio"
	b64 "encoding/base64"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
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

func ParseVariables(data string, agent string, runner string, volume string, vm string, image string, docker string) (outAgent string, executors string) {
	theReader := strings.NewReader(data)
	scanner := bufio.NewScanner(theReader)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, agent) == true {
			outAgent = strings.Replace(line, agent, "", -1)
		}
		if strings.Contains(line, runner) == true {
			executors = strings.Replace(line, runner, "", -1)
		}
		if strings.Contains(line, vm) == true {
			executors = removeText(line, "VM '", "' has been created", -1)
		}
		if strings.Contains(line, volume) == true {
			executors = strings.Replace(line, volume, "", -1)
		}
		if strings.Contains(line, image) == true {
			executors = strings.Replace(line, image, "", -1)
		}
		if strings.Contains(line, docker) == true {
			executors = strings.Replace(line, docker, "", -1)
		}
	}

	return outAgent, executors
}

func GetInput(prompt string) string {
	fmt.Print(prompt)

	// Catch a ^C interrupt.
	// Make sure that we reset term echo before exiting.
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)
	go func() {
		for _ = range signalChannel {
			fmt.Println("\n^C interrupt.")
			termEcho(true)
			os.Exit(1)
		}
	}()

	// Echo is disabled, now grab the data.
	termEcho(false) // disable terminal echo
	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	termEcho(true) // always re-enable terminal echo
	fmt.Println("")
	if err != nil {
		// The terminal has been reset, go ahead and exit.
		fmt.Println("ERROR:", err.Error())
		os.Exit(1)
	}

	sEnc := b64.StdEncoding.EncodeToString([]byte(text))

	return strings.TrimSpace(sEnc)
}

func SetInput() (string, string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Host: ")
	host, _ := reader.ReadString('\n')

	fmt.Print("Namespace: ")
	project_url, _ := reader.ReadString('\n')

	return strings.TrimSpace(host), strings.TrimSpace(project_url)
}

func LetsDecrypt(p string) string {
	sDec, _ := b64.StdEncoding.DecodeString(p)
	return string(sDec)
}

func termEcho(on bool) {
	// Common settings and variables for both stty calls.
	attrs := syscall.ProcAttr{
		Dir:   "",
		Env:   []string{},
		Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd()},
		Sys:   nil}
	var ws syscall.WaitStatus
	cmd := "echo"
	if on == false {
		cmd = "-echo"
	}

	// Enable/disable echoing.
	pid, err := syscall.ForkExec(
		"/bin/stty",
		[]string{"stty", cmd},
		&attrs)
	if err != nil {
		panic(err)
	}

	// Wait for the stty process to complete.
	_, err = syscall.Wait4(pid, &ws, 0, nil)
	if err != nil {
		panic(err)
	}
}
