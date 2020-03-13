package transport

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/hugorut/coins-oracle/pkg/test"
)

var _ = Describe("Client", func() {
	var (
		mockServer *test.Server
	)

	BeforeEach(func() {
		mockServer = test.NewTestServer(GinkgoT())
	})

	AfterEach(func() {
		mockServer.Close()
	})

	Describe("BaseClient", func() {
		var (
			baseClient BaseClient
		)

		type testout struct {
			Data string `json:"data"`
		}

		BeforeEach(func() {
			u, err := url.Parse(mockServer.HttpTest.URL)
			Expect(err).ToNot(HaveOccurred())

			baseClient = BaseClient{
				BaseURL: u,
				Client:  &http.Client{Timeout: DefaultClientTimeout},
				Log:     log.New(ioutil.Discard, "", log.LstdFlags),
			}
		})

		Describe("#GET", func() {
			It("Should execute a GET request and marshall success", func() {
				var out testout

				mockServer.Expect(test.ExpectedCall{
					Path:   "/test/g",
					Method: http.MethodGet,
					QueryParams: map[string]string{
						"param": "1",
					},
					Response:     `{"data": "hello world"}`,
					ResponseCode: http.StatusOK,
				})

				err := baseClient.GET("/test/g", map[string]string{
					"param": "1",
				}, &out)
				Expect(err).ToNot(HaveOccurred())

				Expect(out.Data).To(Equal("hello world"))
			})
		})

		Describe("#POST", func() {
			It("Should execute a POST request and marshall success", func() {
				var out testout

				mockServer.Expect(test.ExpectedCall{
					Path:   "/test/p",
					Method: http.MethodPost,
					Headers: map[string]string{
						"Content-type": "application/json",
					},
					Body:         `{"data": "request"}`,
					Response:     `{"data": "hello world"}`,
					ResponseCode: http.StatusOK,
				})

				err := baseClient.POST(testout{Data: "request"}, "/test/p", &out)
				Expect(err).ToNot(HaveOccurred())

				Expect(out.Data).To(Equal("hello world"))
			})
		})
	})

})
