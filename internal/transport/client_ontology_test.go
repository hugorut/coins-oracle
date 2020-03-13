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

var _ = Describe("OntologyClient", func() {
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

		client = &OntologyClient{
			BaseClient: transport.BaseClient{
				BaseURL: u,
				Client:  http.DefaultClient,
			},
		}
	})

	AfterEach(func() {
		mockServer.Close()
	})

	Describe("#GetInfo", func() {
		It("Should return the Ontology node information transformed to the common output", func() {
			blockNumber := 6810623
			bestBlockHash := "5b24c6d342f527adbce455d69970be823ce81a533109c4723c73716575322d1c"

			mockServer.Expect(test.ExpectedCall{
				Path:         "/api/v1/block/height",
				Method:       http.MethodGet,
				Response:     MustLoad(fb.LoadFixture("ontology/res/getinfo.json", blockNumber)),
				ResponseCode: http.StatusOK,
			}).Then(test.ExpectedCall{
				Path:         "/api/v1/block/hash/6810623",
				Method:       http.MethodGet,
				Response:     MustLoad(fb.LoadFixture("ontology/res/getblockhash.json", bestBlockHash)),
				ResponseCode: http.StatusOK,
			})

			info, err := client.GetInfo()
			Expect(err).ToNot(HaveOccurred())

			Expect(info).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Chain":        Equal("main"),
					"CurrentBlock": Equal(bestBlockHash),
					"BlockHeight":  Equal(blockNumber),
				}),
			})))
		})
	})

	Describe("#GetBalance", func() {
		It("Should return the Ontology balance transformed to the common output", func() {
			addr := "address"
			balRes := "144768442"

			mockServer.Expect(test.ExpectedCall{
				Path:         "/api/v1/balance/address",
				Method:       http.MethodGet,
				Response:     MustLoad(fb.LoadFixture("ontology/res/getbalance.json", balRes)),
				ResponseCode: http.StatusOK,
			})

			balance, err := client.GetBalance(addr)
			Expect(err).ToNot(HaveOccurred())

			Expect(balance).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Assets": ConsistOf(
						MatchAllFields(Fields{
							"Asset":   Equal("ONT"),
							"Balance": Equal(balRes),
						}),
					),
				}),
			})))
		})
	})

	Describe("#GetTransactionByHash", func() {
		It("Should return the Ontology transaction transformed to the common transaction interface", func() {
			txID := "8cbc907de63a58d864606901fd5edab546c31584e2d577962a8ce6cbdce09d92"

			mockServer.Expect(test.ExpectedCall{
				Path:         "/api/v1/transaction/8cbc907de63a58d864606901fd5edab546c31584e2d577962a8ce6cbdce09d92",
				Method:       http.MethodGet,
				Response:     MustLoad(fb.LoadFixture("ontology/res/gettransaction.json", txID)),
				ResponseCode: http.StatusOK,
			})

			tx, err := client.GetTransactionByHash(txID)
			Expect(err).ToNot(HaveOccurred())

			Expect(tx).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Transaction": MatchAllFields(Fields{
						"ID":    Equal(txID),
						"From":  Equal("AFmseVrdL9f9oyCzZefL9tG6UbvhUMqNMV"),
						"To":    Equal("AWM9vmGpAhFyiXxg8r5Cx4H3mS2zrtSkUF"),
						"Value": Equal("1165402205"),
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
