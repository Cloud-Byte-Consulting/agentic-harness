package pack_test

import (
	"os"
	"path/filepath"

	"github.com/cloud-byte/air/internal/pack"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func writePodPack(root, id string) {
	dir := filepath.Join(root, "pod-bundle", "personas", id)
	Expect(os.MkdirAll(dir, 0o755)).To(Succeed())
	for _, f := range pack.RequiredPodFiles {
		Expect(os.WriteFile(filepath.Join(dir, f), []byte("x"), 0o644)).To(Succeed())
	}
}

var _ = Describe("Package", func() {
	It("produces persona tarballs, a bundle, and a checksummed manifest", func() {
		root := GinkgoT().TempDir()
		writePodPack(root, "alpha")
		writePodPack(root, "beta")
		Expect(os.WriteFile(filepath.Join(root, "harness.manifest.yaml"),
			[]byte("harness:\n  name: AIR\n  release: \"9.9\"\ncomponents:\n  - id: x\n"), 0o644)).To(Succeed())

		out := filepath.Join(root, "dist")
		rep, err := pack.Package(root, out)
		Expect(err).NotTo(HaveOccurred())
		Expect(rep.Release).To(Equal("9.9"))
		Expect(rep.Personas).To(HaveLen(2))
		Expect(rep.Bundle.SHA256).To(HaveLen(64))
		for _, p := range []string{
			"personas/alpha.tar.gz",
			"personas/beta.tar.gz",
			"air-harness-9.9.tar.gz",
			"manifest.json",
		} {
			Expect(filepath.Join(out, p)).To(BeAnExistingFile())
		}
	})

	It("fails when a persona pack is incomplete", func() {
		root := GinkgoT().TempDir()
		dir := filepath.Join(root, "pod-bundle", "personas", "broken")
		Expect(os.MkdirAll(dir, 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(dir, "persona.yaml"), []byte("x"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "harness.manifest.yaml"),
			[]byte("harness:\n  release: \"1\"\ncomponents:\n  - id: x\n"), 0o644)).To(Succeed())

		_, err := pack.Package(root, filepath.Join(root, "dist"))
		Expect(err).To(HaveOccurred())
	})
})
