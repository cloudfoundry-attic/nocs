package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"syscall"

	"github.com/cf-guardian/specs"
)

func main() {
	if os.Args[1] != "exec" {
		panic("Unsupported command!")
	}
	config := parseConfig(configFilePath())

	cmd := exec.Command(config.Process.Args[0], config.Process.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	checkError(cmd.Start(), "starting process", 80)

	checkWaitError(cmd.Wait(), "awaiting process completion", 85)
}

func configFilePath() string {
	if len(os.Args) < 3 || os.Args[2] == "" {
		pwd, err := os.Getwd()
		checkError(err, "getting working directory", 90)
		return path.Join(pwd, "config.json")
	}
	return os.Args[2]
}

func parseConfig(configPath string) *specs.Spec {
	configStr, err := ioutil.ReadFile(configPath)
	checkError(err, "reading config file", 95)

	var config = &specs.Spec{}
	checkError(json.Unmarshal([]byte(configStr), config), "parsing config JSON", 100)

	return config
}

func checkWaitError(err error, action string, exitCode int) {
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if waitStatus, ok := exitErr.Sys().(syscall.WaitStatus); !ok {
				checkError(err, action+": expected a WaitStatus", exitCode)
			} else {
				os.Exit(waitStatus.ExitStatus())
			}
		}
		checkError(err, action+": expected an ExitError", exitCode)
	}
}

func checkError(err error, action string, exitCode int) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s: %s\n", filepath.Base(os.Args[0]), action, err)
		os.Exit(exitCode)
	}
}
