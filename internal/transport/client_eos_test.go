package transport_test

import (
	"github.com/hugorut/coins-oracle/pkg/transport"
	"net/http"
	"os"
	"path"

	eos "github.com/eoscanada/eos-go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	. "github.com/hugorut/coins-oracle/internal/transport"
	"github.com/hugorut/coins-oracle/pkg/test"
)

var _ = Describe("EosClient", func() {
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
		client = &EosClient{
			Client: eos.New(mockServer.HttpTest.URL),
		}
	})

	AfterEach(func() {
		mockServer.Close()
	})

	Describe("#GetBalance", func() {
		It("Should return the Eos balance transformed to the common output", func() {
			addr := "eospaceioeos"
			balR := "123"

			mockServer.Expect(test.ExpectedCall{
				Path:         "/v1/chain/get_currency_balance",
				Method:       "POST",
				Body:         MustLoad(fb.LoadFixture("eos/req/getcurrencybalance.json", addr)),
				Response:     MustLoad(fb.LoadFixture("eos/res/getcurrencybalance.json", balR)),
				ResponseCode: http.StatusOK,
			})

			balance, err := client.GetBalance(addr)
			Expect(err).ToNot(HaveOccurred())

			Expect(balance).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Assets": ConsistOf(
						MatchAllFields(Fields{
							"Asset":   Equal("EOS"),
							"Balance": Equal(balR),
						}),
					),
				}),
			})))
		})
	})

	Describe("#GetInfo", func() {
		It("Should return the Eos node information transformed to the common output", func() {
			bestBlockHash := "0024d87f2415c674f166389ba9abba4152053ac8100767ac4a06d9a8c4ab905a"

			mockServer.Expect(test.ExpectedCall{
				Path:         "/v1/chain/get_info",
				Method:       "POST",
				Response:     MustLoad(fb.LoadFixture("eos/res/getinfo.json", bestBlockHash)),
				ResponseCode: http.StatusOK,
			})

			info, err := client.GetInfo()
			Expect(err).ToNot(HaveOccurred())

			Expect(info).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Chain":        Equal("cf057bbfb72640471fd910bcb67639c22df9f92470936cddc1ade0e2f2e7dc4f"),
					"CurrentBlock": Equal(bestBlockHash),
					"BlockHeight":  Equal(21098590),
				}),
			})))
		})
	})

	Describe("#GetTransactionByHash", func() {
		It("Should return the Eos transaction transformed to the common transaction interface", func() {
			txID := "e6c814f9ba58e2aedd654abfdefc99c98f3e4bf5f20e4820b7d212f38f1f6f13"
			bestBlockHash := "0024d87f2415c674f166389ba9abba4152053ac8100767ac4a06d9a8c4ab905a"

			mockServer.Expect(test.ExpectedCall{
				Path:         "/v1/history/get_transaction",
				Method:       "POST",
				Body:         MustLoad(fb.LoadFixture("eos/req/gettransaction.json", txID)),
				Response:     MustLoad(fb.LoadFixture("eos/res/gettransaction.json", txID)),
				ResponseCode: http.StatusOK,
			}).Then(test.ExpectedCall{
				Path:         "/v1/chain/get_info",
				Method:       "POST",
				Response:     MustLoad(fb.LoadFixture("eos/res/getinfo.json", bestBlockHash)),
				ResponseCode: http.StatusOK,
			})

			tx, err := client.GetTransactionByHash(txID)
			Expect(err).ToNot(HaveOccurred())

			Expect(tx).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Transaction": MatchAllFields(Fields{
						"ID":    Equal(txID),
						"From":  Equal("cryptkeeper"),
						"To":    Equal("brandon"),
						"Value": Equal("42.0000 EOS"),
						"Confirmations": MatchAllFields(Fields{
							"Threshold": PointTo(Equal(int64(5))),
							"Confirmed": BeTrue(),
							"Value":     PointTo(Equal(int64(15))),
						}),
					}),
				}),
			})))
		})
	})
})
