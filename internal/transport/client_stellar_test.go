package transport_test

import (
	"github.com/hugorut/coins-oracle/pkg/transport"
	"net/http"
	"os"
	"path"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"github.com/stellar/go/clients/horizonclient"

	. "github.com/hugorut/coins-oracle/internal/transport"
	"github.com/hugorut/coins-oracle/pkg/test"
)

var _ = Describe("StellarClient", func() {
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

		client = &StellarClient{
			Client: &horizonclient.Client{
				HorizonURL: mockServer.HttpTest.URL,
				HTTP:       http.DefaultClient,
			},
		}
	})

	AfterEach(func() {
		mockServer.Close()
	})

	Describe("#GetInfo", func() {
		It("Should return the Stellar node information transformed to the common output", func() {
			bestBlockHash := "15040c1d611cdc88946b47baf19ae78ab1d842487b0e61586de3848065bd3f45"

			mockServer.Expect(test.ExpectedCall{
				Path:   "/ledgers",
				Method: http.MethodGet,
				QueryParams: map[string]string{
					"limit": "1",
					"order": "desc",
				},
				Response:     MustLoad(fb.LoadFixture("stellar/res/getinfo.json", bestBlockHash)),
				ResponseCode: http.StatusOK,
			})

			info, err := client.GetInfo()
			Expect(err).ToNot(HaveOccurred())

			Expect(info).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Chain":        Equal("main"),
					"CurrentBlock": Equal(bestBlockHash),
					"BlockHeight":  Equal(26187596),
				}),
			})))
		})
	})

	Describe("#GetBalance", func() {
		It("Should return the Stellar balance transformed to the common output", func() {
			addr := "GA4MTF3WRJE7I6TSP66PYXBEIS3PPUZEHJYTBMBNAPBBL3TWQWLAZZDW"

			mockServer.Expect(test.ExpectedCall{
				Path:         "/accounts/" + addr,
				Method:       http.MethodGet,
				Response:     MustLoad(fb.LoadFixture("stellar/res/getbalance.json")),
				ResponseCode: http.StatusOK,
			})

			balance, err := client.GetBalance(addr)
			Expect(err).ToNot(HaveOccurred())

			Expect(balance).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Assets": ConsistOf(
						MatchAllFields(Fields{
							"Asset":   Equal("BTC"),
							"Balance": Equal("0.0000000"),
						}),
						MatchAllFields(Fields{
							"Asset":   Equal("NRV"),
							"Balance": Equal("0.0000000"),
						}),
						MatchAllFields(Fields{
							"Asset":   Equal("ETH"),
							"Balance": Equal("0.0000000"),
						}),
						MatchAllFields(Fields{
							"Asset":   Equal("XLM"),
							"Balance": Equal("249.6635636"),
						}),
					),
				}),
			})))
		})
	})

	Describe("#GetTransactionByHash", func() {
		It("Should return the Stellar transaction transformed to the common transaction interface", func() {
			txID := "1124f249dbf509b2cd22a39595967b6ebb4cfb75d99e405d80b0e0006615f2a3"

			mockServer.Expect(test.ExpectedCall{
				Path:         "/transactions/" + txID,
				Method:       http.MethodGet,
				Response:     MustLoad(fb.LoadFixture("stellar/res/gettransaction.json", txID, txID)),
				ResponseCode: http.StatusOK,
			}).Then(test.ExpectedCall{
				Path:         "/transactions/" + txID + "/operations",
				Method:       http.MethodGet,
				Response:     MustLoad(fb.LoadFixture("stellar/res/operations.json")),
				ResponseCode: http.StatusOK,
			})

			tx, err := client.GetTransactionByHash(txID)
			Expect(err).ToNot(HaveOccurred())

			Expect(tx).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Transaction": MatchAllFields(Fields{
						"ID":    Equal(txID),
						"From":  Equal("GDC35NCQORRH7DDIZ4GHR4OB3V7B35V7CRRIWXTUYNKJFBFPUTBPSCOE"),
						"To":    Equal("GBF2RYH7OJOW63HI3CCIF5R7EPK257A3EN6ILH5OGUCJIMR4Z23U6P5V"),
						"Value": Equal("5.0000000"),
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
