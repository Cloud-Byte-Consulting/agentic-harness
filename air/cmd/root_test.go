package cmd

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"truenas-scale-1.tail5a208d.ts.net/Cloud-Byte-Consulting/teo"
)

var _ = Describe("air CLI", func() {
	Describe("status", func() {
		It("lists the manifest components (human)", func() {
			out, err := run("status", "--human", "--manifest", mpath())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("release test-1.0"))
			Expect(out).To(ContainSubstring("router:role-router"))
			Expect(out).To(ContainSubstring("proxy:cachy"))
		})
		It("errors when the manifest is missing", func() {
			_, err := run("status", "--manifest", "nope.yaml")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("persona", func() {
		It("prints a persona pack", func() {
			out, err := run("persona", "software-engineer", "--manifest", mpath())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("role: SWE"))
		})
		It("errors for an unknown persona", func() {
			_, err := run("persona", "no-such-persona", "--manifest", mpath())
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("version", func() {
		It("prints the air version (human)", func() {
			out, err := run("version", "--human", "--manifest", mpath())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("air "))
		})
		It("defaults to TEO", func() {
			out, err := run("version", "--manifest", mpath())
			Expect(err).NotTo(HaveOccurred())
			doc, perr := teo.Parse(out)
			Expect(perr).NotTo(HaveOccurred())
			Expect(doc.GetScalar("air")).NotTo(BeNil())
		})
	})

	Describe("skills", func() {
		It("lists canonical skills", func() {
			out, err := run("skills", "list", "--manifest", mpath())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("example-skill"))
		})
		It("syncs copies into a destination", func() {
			dest := GinkgoT().TempDir()
			_, err := run("skills", "sync", "--manifest", mpath(), "--out", dest)
			Expect(err).NotTo(HaveOccurred())
			for _, p := range []string{
				filepath.Join(".claude", "skills", "example-skill", "SKILL.md"),
				filepath.Join(".cursor", "rules", "example-skill.mdc"),
				filepath.Join(".github", "instructions", "example-skill.instructions.md"),
				filepath.Join(".gemini", "skills", "example-skill.md"),
			} {
				Expect(filepath.Join(dest, p)).To(BeAnExistingFile())
			}
		})
		It("links the Claude skills directory", func() {
			dest := GinkgoT().TempDir()
			_, err := run("skills", "link", "--manifest", mpath(), "--out", dest)
			Expect(err).NotTo(HaveOccurred())
			fi, err := os.Lstat(filepath.Join(dest, ".claude", "skills"))
			Expect(err).NotTo(HaveOccurred())
			Expect(fi.Mode() & os.ModeSymlink).NotTo(BeZero())
		})
	})

	Describe("agents", func() {
		It("fans AGENTS.md out in a project dir", func() {
			proj := GinkgoT().TempDir()
			Expect(os.WriteFile(filepath.Join(proj, "AGENTS.md"), []byte("rules"), 0o644)).To(Succeed())
			_, err := run("agents", "link", "--root", proj, "--harness", "claude")
			Expect(err).NotTo(HaveOccurred())
			Expect(filepath.Join(proj, "CLAUDE.md")).To(BeAnExistingFile())
		})
	})

	// Every --teo output must parse as valid TEO and reconstruct the expected shape.
	Describe("TEO conformance", func() {
		parse := func(args ...string) *teo.Document {
			out, err := run(append(args, "--teo", "--manifest", mpath())...)
			Expect(err).NotTo(HaveOccurred())
			doc, perr := teo.Parse(out)
			Expect(perr).NotTo(HaveOccurred(), "output must be valid TEO:\n"+out)
			return doc
		}

		It("status emits a typed components block", func() {
			doc := parse("status")
			Expect(doc.GetScalar("count")).To(Equal(4))
			blk := doc.FindBlock("components")
			Expect(blk).NotTo(BeNil())
			Expect(blk.Fields).To(Equal([]string{"id", "kind", "layer", "lifecycle"}))
			Expect(blk.Rows).To(HaveLen(4))
		})

		It("personas list emits a personas block", func() {
			doc := parse("personas", "list")
			blk := doc.FindBlock("personas")
			Expect(blk).NotTo(BeNil())
			Expect(blk.Fields).To(Equal([]string{"id", "role", "gate"}))
		})

		It("targets list emits harnesses and sets blocks", func() {
			doc := parse("targets", "list")
			Expect(doc.FindBlock("harnesses")).NotTo(BeNil())
			Expect(doc.FindBlock("sets")).NotTo(BeNil())
		})

		It("version emits scalars", func() {
			doc := parse("version")
			Expect(doc.GetScalar("air")).NotTo(BeNil())
		})
	})
})
