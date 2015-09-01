package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"

	"github.com/onsi/gomega/gexec"
)

var nocsBin string

func TestNocs(t *testing.T) {
	RegisterFailHandler(Fail)

	SynchronizedBeforeSuite(func() []byte {
		nocsPath, err := gexec.Build("github.com/cloudfoundry-incubator/nocs")
		Expect(err).ToNot(HaveOccurred())
		return []byte(nocsPath)
	}, func(path []byte) {
		Expect(string(path)).NotTo(BeEmpty())
		nocsBin = string(path)
	})

	RunSpecs(t, "nOCS Suite")
}
