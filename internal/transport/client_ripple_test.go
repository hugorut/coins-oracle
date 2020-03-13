package transport_test

import (
	"github.com/hugorut/coins-oracle/pkg/transport"
	"net/http"
	"net/url"
	"os"
	"path"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	. "github.com/hugorut/coins-oracle/internal/transport"
	"github.com/hugorut/coins-oracle/pkg/test"
)

var _ = Describe("RippleClient", func() {
	var (
		fb     *test.FixtureBox
		client transport.CoinClient

		mockServer *test.Server
	)

	BeforeEach(func() {
		dir, err := os.Getwd()
		Expect(err).ToNot(HaveOccurred())

		fb = &test.FixtureBox{
			Base: path.Join(dir, "../test/fixtures"),
		}

		mockServer = test.NewTestServer(GinkgoT())
		u, err := url.Parse(mockServer.HttpTest.URL)
		Expect(err).ToNot(HaveOccurred())

		client = &RippleClient{
			BaseClient: transport.BaseClient{
				BaseURL: u,
				Client:  http.DefaultClient,
			},
		}
	})

	AfterEach(func() {
		mockServer.Close()
	})

	Describe("#GetBalance", func() {
		It("Should return the Ripple balance transformed to the common output", func() {
			addr := "rLgBm6vum6YLS3j88Cv7F27pR3FbJssuph"
			balRes := "153881000"

			mockServer.Expect(test.ExpectedCall{
				Path:   "/",
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "Application/Json",
				},
				Body:         MustLoad(fb.LoadFixture("ripple/req/getbalance.json", addr)),
				Response:     MustLoad(fb.LoadFixture("ripple/res/getbalance.json", addr, balRes)),
				ResponseCode: http.StatusOK,
			})

			balance, err := client.GetBalance(addr)
			Expect(err).ToNot(HaveOccurred())

			Expect(balance).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Assets": ConsistOf(
						MatchAllFields(Fields{
							"Asset":   Equal("XRP"),
							"Balance": Equal("153.881"),
						}),
					),
				}),
			})))
		})
	})

	Describe("#GetInfo", func() {
		It("Should return the Ripple node information transformed to the common output", func() {
			bestBlockHash := "329BDAFA8F11D4878EE03BADFA723EA577A66D5483AAF80EF5DE4C63012162AB"

			mockServer.Expect(test.ExpectedCall{
				Path:   "/",
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "Application/Json",
				},
				Body:         MustLoad(fb.LoadFixture("ripple/req/getinfo.json")),
				Response:     MustLoad(fb.LoadFixture("ripple/res/getinfo.json", bestBlockHash)),
				ResponseCode: http.StatusOK,
			})

			info, err := client.GetInfo()
			Expect(err).ToNot(HaveOccurred())

			Expect(info).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Chain":        Equal("main"),
					"BlockHeight":  Equal(50411128),
					"CurrentBlock": Equal(bestBlockHash),
				}),
			})))
		})
	})

	Describe("#GetTransactionByHash", func() {
		It("Should return the Ripple transaction transformed to the common transaction interface", func() {
			txID := "4A12A8759149C2888B8AFCCF7B5C0423D3BBA2EF72F4D8672182601301A4F798"

			mockServer.Expect(test.ExpectedCall{
				Path:   "/",
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "Application/Json",
				},
				Body:         MustLoad(fb.LoadFixture("ripple/req/gettransaction.json", txID)),
				Response:     MustLoad(fb.LoadFixture("ripple/res/gettransaction.json", txID)),
				ResponseCode: http.StatusOK,
			})

			tx, err := client.GetTransactionByHash(txID)
			Expect(err).ToNot(HaveOccurred())

			Expect(tx).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Transaction": MatchAllFields(Fields{
						"ID":    Equal(txID),
						"From":  Equal("rP1afBEfikTz7hJh2ExCDni9W4Bx1dUMRk"),
						"To":    Equal("rMZdHB6uHvAEPzzKdsWYyhgyLkhFNjuwih"),
						"Value": Equal("206.17"),
						"Confirmations": MatchAllFields(Fields{
							"Threshold": BeNil(),
							"Confirmed": BeTrue(),
							"Value":     BeNil(),
						}),
					}),
				}),
			})))
		})

		Context("With amount as object", func() {
			It("Should return value in correct format", func() {
				txID := "4A12A8759149C2888B8AFCCF7B5C0423D3BBA2EF72F4D8672182601301A4F798"

				mockServer.Expect(test.ExpectedCall{
					Path:   "/",
					Method: "POST",
					Headers: map[string]string{
						"Content-Type": "Application/Json",
					},
					Body:         MustLoad(fb.LoadFixture("ripple/req/gettransaction.json", txID)),
					Response:     MustLoad(fb.LoadFixture("ripple/res/gettransaction_objectamount.json", txID)),
					ResponseCode: http.StatusOK,
				})

				tx, err := client.GetTransactionByHash(txID)
				Expect(err).ToNot(HaveOccurred())

				Expect(tx).To(PointTo(MatchAllFields(Fields{
					"Data": MatchAllFields(Fields{
						"Transaction": MatchAllFields(Fields{
							"ID":    Equal(txID),
							"From":  Equal("rf3B8KcYqKMgybB2ms9KcLhcB8bWX1UDov"),
							"To":    Equal("rf3B8KcYqKMgybB2ms9KcLhcB8bWX1UDov"),
							"Value": Equal("0.005001"),
							"Confirmations": MatchAllFields(Fields{
								"Threshold": BeNil(),
								"Confirmed": BeTrue(),
								"Value":     BeNil(),
							}),
						}),
					}),
				})))
			})
		})
	})
})
