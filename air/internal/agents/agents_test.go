package agents_test

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/cloud-byte/air/internal/agents"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Agents", func() {
	Describe("ResolveMode", func() {
		It("maps auto to the OS default", func() {
			want := "symlink"
			if runtime.GOOS == "windows" {
				want = "copy"
			}
			Expect(agents.ResolveMode("auto")).To(Equal(want))
		})
		It("passes explicit modes through", func() {
			Expect(agents.ResolveMode("copy")).To(Equal("copy"))
			Expect(agents.ResolveMode("symlink")).To(Equal("symlink"))
		})
	})

	Describe("ResolveTargets", func() {
		It("skips the harness whose path equals the canonical", func() {
			got, err := agents.ResolveTargets(agents.Options{Canonical: "AGENTS.md", All: true})
			Expect(err).NotTo(HaveOccurred())
			for _, h := range got {
				Expect(h.Path).NotTo(Equal("AGENTS.md"))
			}
		})
		It("errors on an unknown harness", func() {
			_, err := agents.ResolveTargets(agents.Options{Canonical: "AGENTS.md", Harnesses: []string{"nope"}})
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Link", func() {
		var root string
		BeforeEach(func() { root = GinkgoT().TempDir() })

		It("writes copies with the canonical content (copy mode)", func() {
			Expect(os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte("canon"), 0o644)).To(Succeed())
			_, err := agents.Link(agents.Options{Root: root, Mode: "copy", Harnesses: []string{"claude", "copilot"}})
			Expect(err).NotTo(HaveOccurred())
			body, err := os.ReadFile(filepath.Join(root, "CLAUDE.md"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(body)).To(Equal("canon"))
			Expect(filepath.Join(root, ".github", "copilot-instructions.md")).To(BeAnExistingFile())
		})

		It("creates a symlink and is idempotent on re-run", func() {
			Expect(os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte("canon"), 0o644)).To(Succeed())
			_, err := agents.Link(agents.Options{Root: root, Mode: "symlink", Harnesses: []string{"gemini"}})
			Expect(err).NotTo(HaveOccurred())
			fi, err := os.Lstat(filepath.Join(root, "GEMINI.md"))
			Expect(err).NotTo(HaveOccurred())
			Expect(fi.Mode() & os.ModeSymlink).NotTo(BeZero())

			rep, err := agents.Link(agents.Options{Root: root, Mode: "symlink", Harnesses: []string{"gemini"}})
			Expect(err).NotTo(HaveOccurred())
			Expect(rep.Actions).To(ContainElement(HaveField("Verb", "ok")))
		})

		It("backs up an existing real file to .bak", func() {
			os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte("canon"), 0o644)
			os.WriteFile(filepath.Join(root, "CLAUDE.md"), []byte("pre-existing"), 0o644)
			_, err := agents.Link(agents.Options{Root: root, Mode: "copy", Harnesses: []string{"claude"}})
			Expect(err).NotTo(HaveOccurred())
			bak, err := os.ReadFile(filepath.Join(root, "CLAUDE.md.bak"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(bak)).To(Equal("pre-existing"))
		})

		It("seeds a starter AGENTS.md when missing", func() {
			_, err := agents.Link(agents.Options{Root: root, Mode: "copy", Harnesses: []string{"claude"}})
			Expect(err).NotTo(HaveOccurred())
			Expect(filepath.Join(root, "AGENTS.md")).To(BeAnExistingFile())
		})

		It("changes nothing in dry-run", func() {
			os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte("canon"), 0o644)
			_, err := agents.Link(agents.Options{Root: root, Mode: "symlink", Harnesses: []string{"claude"}, DryRun: true})
			Expect(err).NotTo(HaveOccurred())
			Expect(filepath.Join(root, "CLAUDE.md")).NotTo(BeAnExistingFile())
		})
	})
})
