package personas

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPersonas(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Personas Suite")
}
