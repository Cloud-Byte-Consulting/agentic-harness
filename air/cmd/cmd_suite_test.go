package cmd

import (
	"bytes"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
)

func TestCmd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cmd Suite")
}

func mpath() string { return filepath.Join("..", "testdata", "harness.manifest.yaml") }

// run executes the root command with a clean viper state and captures output.
func run(args ...string) (string, error) {
	viper.Reset()
	root := NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(args)
	err := root.Execute()
	return buf.String(), err
}
