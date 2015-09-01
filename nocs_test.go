package main_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"

	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("nOCS", func() {
	var (
		configFileDir string
		nocsProcessWD string

		cmd *exec.Cmd

		stdout, stderr string
		exitCode       int
	)

	BeforeEach(func() {
		dirtyPath, err := ioutil.TempDir("", "nOCSTest")
		Expect(err).NotTo(HaveOccurred())
		configFileDir, err = filepath.EvalSymlinks(dirtyPath)
		Expect(err).NotTo(HaveOccurred())

		nocsProcessWD, err = os.Getwd()
		Expect(err).NotTo(HaveOccurred())
	})

	JustBeforeEach(func() {
		cmd.Dir = nocsProcessWD

		outBuffer, errBuffer := bytes.NewBuffer([]byte{}), bytes.NewBuffer([]byte{})
		cmd.Stdout, cmd.Stderr = outBuffer, errBuffer

		Expect(cmd.Start()).To(Succeed())

		exitCode = getExitCode(cmd.Wait())

		stdout, stderr = outBuffer.String(), errBuffer.String()
	})

	AfterEach(func() {
		os.RemoveAll(configFileDir)
	})

	Context("Running a simple command", func() {
		BeforeEach(func() {
			commandArgs := []string{"echo", "Hello OCF!"}
			configFilePath := createConfigFile("nocs-simple.json", configFileDir, commandArgs)

			cmd = exec.Command(nocsBin, "exec", configFilePath)
		})

		It("produces the appropriate stdout", func() {
			Expect(stdout).To(Equal("Hello OCF!\n"))
		})

		It("produces the appropriate stderr", func() {
			Expect(stderr).To(BeEmpty())
		})

		It("returns the appropriate exit status code", func() {
			Expect(exitCode).To(Equal(0))
		})
	})

	Context("Running a command which produces stderr", func() {
		BeforeEach(func() {
			commandArgs := []string{"/bin/sh", "-c", "echo Hello OCF! 1>&2"}
			configFilePath := createConfigFile("nocs-stderr.json", configFileDir, commandArgs)

			cmd = exec.Command(nocsBin, "exec", configFilePath)
		})

		It("produces the appropriate stdout", func() {
			Expect(stdout).To(BeEmpty())
		})

		It("produces the appropriate stderr", func() {
			Expect(stderr).To(Equal("Hello OCF!\n"))
		})

		It("returns the appropriate exit status code", func() {
			Expect(exitCode).To(Equal(0))
		})
	})

	Context("Running a command which exits with non-zero status", func() {
		BeforeEach(func() {
			commandArgs := []string{"/usr/bin/false"}
			configFilePath := createConfigFile("nocs-errcode.json", configFileDir, commandArgs)

			cmd = exec.Command(nocsBin, "exec", configFilePath)
		})

		It("returns the appropriate exit status code", func() {
			Expect(exitCode).To(Equal(1))
		})
	})

	Context("Running without passing a configuration filepath", func() {
		BeforeEach(func() {
			commandArgs := []string{"echo", "Hello default OCF!"}
			createConfigFile("config.json", configFileDir, commandArgs)

			cmd = exec.Command(nocsBin, "exec")

			nocsProcessWD = configFileDir
		})

		It("produces the appropriate stdout", func() {
			Expect(stdout).To(Equal("Hello default OCF!\n"))
		})
	})

	Context("Running when passing the empty configuration filepath", func() {
		BeforeEach(func() {
			commandArgs := []string{"echo", "Hello empty OCF!"}
			createConfigFile("config.json", configFileDir, commandArgs)

			cmd = exec.Command(nocsBin, "exec", "")

			nocsProcessWD = configFileDir
		})

		It("produces the appropriate stdout", func() {
			Expect(stdout).To(Equal("Hello empty OCF!\n"))
		})
	})

	Context("Running with a missing configuration file", func() {
		BeforeEach(func() {
			cmd = exec.Command(nocsBin, "exec")

			nocsProcessWD = configFileDir
		})

		It("produces the appropriate stderr", func() {
			Expect(stderr).To(Equal(filepath.Base(nocsBin) + ": reading config file: open " + filepath.Join(configFileDir, "config.json: no such file or directory\n")))
		})

		It("returns the appropriate exit status code", func() {
			Expect(exitCode).To(Equal(95))
		})

	})

	Context("Running when passing a relative configuration filepath", func() {
		BeforeEach(func() {
			commandArgs := []string{"echo", "Hello relative OCF!"}
			createConfigFile("relative.json", configFileDir, commandArgs)
			subdirPath := filepath.Join(configFileDir, "some-subdir")
			err := os.Mkdir(subdirPath, 0777)
			Expect(err).NotTo(HaveOccurred())

			cmd = exec.Command(nocsBin, "exec", "../relative.json")

			nocsProcessWD = subdirPath
		})

		It("produces the appropriate stdout", func() {
			Expect(stdout).To(Equal("Hello relative OCF!\n"))
		})
	})
})

func createConfigFile(configFileName, configFileDir string, args []string) string {
	configFilePath := path.Join(configFileDir, configFileName)

	err := ioutil.WriteFile(configFilePath, []byte(
		`{"process": {
        "user": {
          "uid": 1,
          "gid": 1
        },
        "args": `+argsToJSON(args)+`
     }}`), 0777)
	Expect(err).NotTo(HaveOccurred())

	return configFilePath
}

func argsToJSON(args []string) string {
	return `["` + strings.Join(args, `","`) + `"]`
}

func getExitCode(err error) int {
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		Expect(ok).To(BeTrue())

		waitStatus, ok := exitErr.Sys().(syscall.WaitStatus)
		Expect(ok).To(BeTrue())

		return waitStatus.ExitStatus()
	}
	return 0
}
