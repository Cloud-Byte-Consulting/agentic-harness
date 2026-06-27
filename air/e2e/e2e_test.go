//go:build e2e

// Package e2e builds the air binary and exercises it as a black box.
// Run with: go test -tags e2e ./e2e/...
package e2e

import (
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("air binary", Ordered, func() {
	var bin, manifest string

	BeforeAll(func() {
		bin = filepath.Join(GinkgoT().TempDir(), "air")
		build := exec.Command("go", "build", "-o", bin, ".")
		build.Dir = ".." // module root (air/)
		out, err := build.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), string(out))

		abs, err := filepath.Abs(filepath.Join("..", "testdata", "harness.manifest.yaml"))
		Expect(err).NotTo(HaveOccurred())
		manifest = abs
	})

	run := func(args ...string) string {
		out, err := exec.Command(bin, append(args, "--manifest", manifest)...).CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), string(out))
		return string(out)
	}

	It("reports status as TEO by default", func() {
		out := run("status")
		Expect(out).To(ContainSubstring("components[4]{"))
		Expect(out).To(ContainSubstring("test-1.0"))
	})
	It("still emits human output with --human", func() {
		Expect(run("status", "--human")).To(ContainSubstring("release test-1.0"))
	})
	It("prints its version (TEO)", func() {
		Expect(strings.Contains(run("version"), "air:")).To(BeTrue())
	})
})
