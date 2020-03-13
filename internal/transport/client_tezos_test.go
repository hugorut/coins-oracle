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

var _ = Describe("TezosClient", func() {
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

		client = &TezosClient{
			BaseClient: transport.BaseClient{
				BaseURL: u,
				Client:  http.DefaultClient,
			},
			APIClient: transport.BaseClient{
				BaseURL: u,
				Client:  http.DefaultClient,
			},
		}
	})

	AfterEach(func() {
		mockServer.Close()
	})

	Describe("#GetInfo", func() {
		It("Should return the Tezos node information transformed to the common output", func() {
			bestBlockHash := "hash"
			height := 8883

			mockServer.Expect(test.ExpectedCall{
				Path:         "/chains/main/blocks/head",
				Method:       http.MethodGet,
				Response:     MustLoad(fb.LoadFixture("tezos/res/getinfo.json", height, bestBlockHash)),
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
		It("Should return the Tezos balance transformed to the common output", func() {
			addr := "address"
			balRes := "88348"

			mockServer.Expect(test.ExpectedCall{
				Path:         "/chains/main/balance/head/context/contracts/" + addr,
				Method:       http.MethodGet,
				Response:     MustLoad(fb.LoadFixture("tezos/res/getbalance.json", balRes)),
				ResponseCode: http.StatusOK,
			})

			balance, err := client.GetBalance(addr)
			Expect(err).ToNot(HaveOccurred())

			Expect(balance).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Assets": ConsistOf(
						MatchAllFields(Fields{
							"Asset":   Equal("XTZ"),
							"Balance": Equal(balRes),
						}),
					),
				}),
			})))
		})
	})

	Describe("#GetTransactionByHash", func() {
		It("Should return the Tezos transaction transformed to the common transaction interface", func() {
			txID := "id"

			mockServer.Expect(test.ExpectedCall{
				Path:   "/mooncake/mainnet/v1/transactions",
				Method: http.MethodGet,
				QueryParams: map[string]string{
					"op": txID,
				},
				Response:     MustLoad(fb.LoadFixture("tezos/res/gettransaction.json")),
				ResponseCode: http.StatusOK,
			})

			tx, err := client.GetTransactionByHash(txID)
			Expect(err).ToNot(HaveOccurred())

			Expect(tx).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Transaction": MatchAllFields(Fields{
						"ID":    Equal(txID),
						"From":  Equal("tz1eDDuQBEgwvc6tbnnCVrnr12tvrd6gBTpx"),
						"To":    Equal("KT1AzVUwY4Qynq2K6s1A6ZMv2Uch6FeQtYE2"),
						"Value": Equal("11260000"),
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
