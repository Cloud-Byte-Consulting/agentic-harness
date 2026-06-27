package targets_test

import (
	"github.com/cloud-byte/air/internal/targets"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Targets", func() {
	Describe("Resolve", func() {
		It("expands a comma list of harness names", func() {
			got, err := targets.Resolve("claude,copilot", nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal([]string{"claude", "copilot"}))
		})
		It("expands a builtin set", func() {
			got, err := targets.Resolve("popular", nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(ContainElements("claude", "gemini", "copilot", "cursor"))
		})
		It("expands a user-defined set", func() {
			got, err := targets.Resolve("frontend", map[string][]string{"frontend": {"claude", "copilot"}})
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal([]string{"claude", "copilot"}))
		})
		It("dedupes across names and sets", func() {
			got, err := targets.Resolve("claude,popular,claude", nil)
			Expect(err).NotTo(HaveOccurred())
			seen := map[string]int{}
			for _, h := range got {
				seen[h]++
			}
			Expect(seen["claude"]).To(Equal(1))
		})
		It("errors on an unknown token", func() {
			_, err := targets.Resolve("nope", nil)
			Expect(err).To(HaveOccurred())
		})
		It("errors when a set references an unknown harness", func() {
			_, err := targets.Resolve("bad", map[string][]string{"bad": {"not-a-harness"}})
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("SkillsTools", func() {
		It("keeps skills-capable harnesses and drops the rest", func() {
			Expect(targets.SkillsTools([]string{"claude", "windsurf", "copilot", "kiro"})).
				To(Equal([]string{"claude", "copilot"}))
		})
	})

	Describe("LoadSets", func() {
		It("returns nil for a missing file", func() {
			sets, err := targets.LoadSets("/no/such/file.yaml")
			Expect(err).NotTo(HaveOccurred())
			Expect(sets).To(BeNil())
		})
	})
})
