package transport_test

//go:generate mockgen -destination mocks/client_mocks.go -source ./client.go
//go:generate mockgen -destination mocks/resolver_mocks.go -source ./resolver.go

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBooks(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Transport Suite")
}


// MustLoad wraps a load template with a gomega expectation.
func MustLoad(s string, err error) string  {
	Expect(err).ToNot(HaveOccurred())

	return s
}
