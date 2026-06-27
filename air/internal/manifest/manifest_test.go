package manifest

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var fixture = filepath.Join("..", "..", "testdata", "harness.manifest.yaml")

var _ = Describe("Manifest", func() {
	Describe("Load", func() {
		It("parses harness + components", func() {
			m, err := Load(fixture)
			Expect(err).NotTo(HaveOccurred())
			Expect(m.Harness.Release).To(Equal("test-1.0"))
			Expect(m.Components).To(HaveLen(4))
		})

		It("parses component fields", func() {
			m, err := Load(fixture)
			Expect(err).NotTo(HaveOccurred())
			var cachy *Component
			for i := range m.Components {
				if m.Components[i].ID == "urn:air:test:proxy:cachy" {
					cachy = &m.Components[i]
				}
			}
			Expect(cachy).NotTo(BeNil())
			Expect(cachy.Kind).To(Equal("oci"))
			Expect(cachy.Layer).To(Equal("platform"))
			Expect(cachy.Lifecycle).To(Equal("service"))
			Expect(cachy.Action()).To(Equal("fetch+run"))
		})

		It("errors on a missing file", func() {
			_, err := Load("does-not-exist.yaml")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("ServesPersona", func() {
		It(`treats "all" as serving anyone`, func() {
			Expect(Component{Personas: []string{"all"}}.ServesPersona("anything")).To(BeTrue())
		})
		It("matches a listed persona and rejects others", func() {
			c := Component{Personas: []string{"software-engineer"}}
			Expect(c.ServesPersona("software-engineer")).To(BeTrue())
			Expect(c.ServesPersona("data-engineer")).To(BeFalse())
		})
		It("treats an empty persona as no filter", func() {
			Expect(Component{Personas: []string{"software-engineer"}}.ServesPersona("")).To(BeTrue())
		})
	})

	DescribeTable("Action by lifecycle",
		func(lifecycle, want string) {
			Expect(Component{Lifecycle: lifecycle}.Action()).To(Equal(want))
		},
		Entry("service", "service", "fetch+run"),
		Entry("content", "content", "extract"),
		Entry("tool", "tool", "fetch"),
		Entry("empty", "", "fetch"),
	)
})
