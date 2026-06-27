package skills

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const sampleSkill = "---\ndescription: \"Do a thing well.\"\n---\n\n# /demo\n\nbody line one\nbody line two\n"

func writeSkill(root, name, content string) {
	dir := filepath.Join(root, name)
	Expect(os.MkdirAll(dir, 0o755)).To(Succeed())
	Expect(os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644)).To(Succeed())
}

var _ = Describe("Skills", func() {
	Describe("parseFrontmatter", func() {
		It("reads an inline description and body", func() {
			desc, body := parseFrontmatter(sampleSkill)
			Expect(desc).To(Equal("Do a thing well."))
			Expect(body).To(HavePrefix("# /demo"))
			Expect(body).To(ContainSubstring("body line two"))
		})
		It("folds a block-scalar description", func() {
			desc, _ := parseFrontmatter("---\ndescription: >-\n  Folded line one\n  and line two.\n---\n\nbody\n")
			Expect(desc).To(Equal("Folded line one and line two."))
		})
		It("handles a file with no frontmatter", func() {
			desc, body := parseFrontmatter("# no frontmatter\n\ntext")
			Expect(desc).To(BeEmpty())
			Expect(body).To(ContainSubstring("no frontmatter"))
		})
	})

	Describe("Discover", func() {
		It("finds and sorts skill dirs, ignoring non-skill dirs", func() {
			dir := GinkgoT().TempDir()
			writeSkill(dir, "beta", sampleSkill)
			writeSkill(dir, "alpha", sampleSkill)
			Expect(os.MkdirAll(filepath.Join(dir, "notaskill"), 0o755)).To(Succeed())

			got, err := Discover(dir)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(HaveLen(2))
			Expect(got[0].Name).To(Equal("alpha"))
			Expect(got[1].Name).To(Equal("beta"))
			Expect(got[0].Dir).NotTo(BeEmpty())
		})
	})

	Describe("Project", func() {
		s := Skill{Name: "demo", Description: "Do a thing.", Body: "BODY"}
		DescribeTable("per-tool frontmatter",
			func(tool, fmHas string) {
				content, err := Project(tool, s)
				Expect(err).NotTo(HaveOccurred())
				Expect(content).To(ContainSubstring(fmHas))
				Expect(content).To(ContainSubstring("BODY"))
			},
			Entry("cursor", "cursor", "alwaysApply: false"),
			Entry("copilot", "copilot", "applyTo:"),
			Entry("gemini", "gemini", "# demo"),
		)
		It("rejects claude (not file-projectable)", func() {
			_, err := Project("claude", s)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Sync", func() {
		It("copies per-tool files and a full Claude directory", func() {
			src := GinkgoT().TempDir()
			writeSkill(src, "demo", sampleSkill)
			Expect(os.WriteFile(filepath.Join(src, "demo", "REF.md"), []byte("ref"), 0o644)).To(Succeed())
			out := GinkgoT().TempDir()

			counts, err := Sync(src, out, Tools)
			Expect(err).NotTo(HaveOccurred())
			for _, tool := range Tools {
				Expect(counts[tool]).To(Equal(1), tool)
			}
			for _, p := range []string{
				".claude/skills/demo/SKILL.md",
				".claude/skills/demo/REF.md",
				".cursor/rules/demo.mdc",
				".github/instructions/demo.instructions.md",
				".gemini/skills/demo.md",
				".gemini/skills/INDEX.md",
			} {
				Expect(filepath.Join(out, p)).To(BeAnExistingFile())
			}
		})
	})

	Describe("Link", func() {
		It("symlinks every tool view to the canonical source", func() {
			src := GinkgoT().TempDir()
			writeSkill(src, "demo", sampleSkill)
			out := GinkgoT().TempDir()

			_, err := Link(src, out, Tools)
			Expect(err).NotTo(HaveOccurred())
			for _, l := range []string{".claude/skills", ".cursor/rules/demo.mdc", ".github/instructions/demo.instructions.md", ".gemini/skills/demo.md"} {
				fi, err := os.Lstat(filepath.Join(out, l))
				Expect(err).NotTo(HaveOccurred(), l)
				Expect(fi.Mode()&os.ModeSymlink).NotTo(BeZero(), l+" should be a symlink")
			}
			got, err := os.ReadFile(filepath.Join(out, ".claude", "skills", "demo", "SKILL.md"))
			Expect(err).NotTo(HaveOccurred())
			Expect(strings.Contains(string(got), "# /demo")).To(BeTrue())
		})
	})
})
