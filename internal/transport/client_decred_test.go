package transport_test

import (
	"net/http"
	"net/url"
	"os"
	"path"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	. "github.com/hugorut/coins-oracle/internal/transport"
	"github.com/hugorut/coins-oracle/pkg/test"
	. "github.com/hugorut/coins-oracle/pkg/transport"
)

var _ = Describe("DecredClient", func() {
	var (
		fb     *test.FixtureBox
		client CoinClient

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

		client = &DecredClient{
			BaseClient: BaseClient{
				BaseURL: u,
				Client:  http.DefaultClient,
			},
		}
	})

	AfterEach(func() {
		mockServer.Close()
	})

	Describe("#GetInfo", func() {
		It("Should return the Decred node information transformed to the common output", func() {
			height := 2233
			bestBlockHash := "000000000000000004ac4c6281047603edcbcad5cdfbad017a4d16b5abf43999"

			mockServer.Expect(test.ExpectedCall{
				Path:   "/insight/api/blocks",
				Method: http.MethodGet,
				QueryParams: map[string]string{
					"limit": "1",
				},
				Response:     MustLoad(fb.LoadFixture("decred/res/getinfo.json", height, bestBlockHash)),
				ResponseCode: http.StatusOK,
			})

			info, err := client.GetInfo()
			Expect(err).ToNot(HaveOccurred())

			Expect(info).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Chain":        Equal("main"),
					"CurrentBlock": Equal(bestBlockHash),
					"BlockHeight":  Equal(height),
				}),
			})))
		})
	})

	Describe("#GetBalance", func() {
		It("Should return the Decred balance transformed to the common output", func() {
			addr := "adddr"
			balRes := 3500.74025814

			mockServer.Expect(test.ExpectedCall{
				Path:   "/insight/api/addr/" + addr,
				Method: http.MethodGet,
				QueryParams: map[string]string{
					"noTxList": "1",
				},
				Response:     MustLoad(fb.LoadFixture("decred/res/getbalance.json", balRes)),
				ResponseCode: http.StatusOK,
			})

			balance, err := client.GetBalance(addr)
			Expect(err).ToNot(HaveOccurred())

			Expect(balance).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Assets": ConsistOf(
						MatchAllFields(Fields{
							"Asset":   Equal("DCR"),
							"Balance": Equal("3500.740258"),
						}),
					),
				}),
			})))
		})
	})

	Describe("#GetTransactionByHash", func() {
		It("Should return the Decred transaction transformed to the common transaction interface", func() {
			txID := "6fcacfb574c843b742f7809db4a0be69fe458eb0fb9f4d9ff6b653de829fd385"

			mockServer.Expect(test.ExpectedCall{
				Path:   "/insight/api/tx/" + txID,
				Method: http.MethodGet,
				Response:     MustLoad(fb.LoadFixture("decred/res/gettransaction.json", txID)),
				ResponseCode: http.StatusOK,
			})

			tx, err := client.GetTransactionByHash(txID)
			Expect(err).ToNot(HaveOccurred())

			Expect(tx).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Transaction": MatchAllFields(Fields{
						"ID":    Equal(txID),
						"From":  Equal("DseJP5DPT9jGRpM74wAmVLfdp58VrbQ19zV"),
						"To":    Equal("DsT5LpcLxofEfNPZaQew3PQDTHstUk68kLp"),
						"Value": Equal("820.374097"),
						"Confirmations": MatchAllFields(Fields{
							"Threshold": PointTo(Equal(int64(5))),
							"Confirmed": BeTrue(),
							"Value":     PointTo(Equal(int64(8))),
						}),
					}),
				}),
			})))
		})
	})
})
