package personas

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func specByID(id string) Spec {
	for _, s := range Specs {
		if s.ID == id {
			return s
		}
	}
	Fail("no spec " + id)
	return Spec{}
}

var _ = Describe("Personas", func() {
	Describe("Specs", func() {
		It("has 9 complete persona specs", func() {
			Expect(Specs).To(HaveLen(9))
			for _, s := range Specs {
				Expect(s.ID).NotTo(BeEmpty())
				Expect(s.Role).NotTo(BeEmpty())
				Expect(s.Tier).NotTo(BeEmpty())
				Expect(s.Shared).NotTo(BeEmpty())
			}
		})
	})

	Describe("YAML", func() {
		It("renders router and the net-new comment for SRE", func() {
			y := specByID("site-reliability-engineer").YAML()
			Expect(y).To(ContainSubstring("id: site-reliability-engineer"))
			Expect(y).To(ContainSubstring("defaultMutationTier: tier3_high_exposure"))
			Expect(y).To(ContainSubstring("router: [team-alpha]"))
			Expect(y).To(ContainSubstring("# net-new, author under pod-bundle/skills/"))
		})
		It("renders an empty add list with no comment", func() {
			Expect(specByID("software-engineer").YAML()).To(ContainSubstring("add: []"))
		})
	})

	Describe("Scaffold", func() {
		It("stamps the template and writes persona.yaml for every persona", func() {
			tmpl := GinkgoT().TempDir()
			for _, f := range podFiles {
				Expect(os.WriteFile(filepath.Join(tmpl, f), []byte("# __POD_NAME__ (__POD_ID__)\n"), 0o644)).To(Succeed())
			}
			out := GinkgoT().TempDir()

			made, err := Scaffold(tmpl, out)
			Expect(err).NotTo(HaveOccurred())
			Expect(made).To(HaveLen(9))

			pod, err := os.ReadFile(filepath.Join(out, "data-engineer", "pod.md"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(pod)).To(ContainSubstring("Data Engineer (data-engineer)"))
			Expect(filepath.Join(out, "data-engineer", "persona.yaml")).To(BeAnExistingFile())
		})
	})
})
