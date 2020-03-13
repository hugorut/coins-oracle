package transport_test

import (
	"fmt"
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

var _ = Describe("WavesClient", func() {
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

		client = &WavesClient{
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
		It("Should return the Waves node information transformed to the common output", func() {
			bestBlockHash := 11042

			mockServer.Expect(test.ExpectedCall{
				Path:         "/blocks/last",
				Method:       http.MethodGet,
				Response:     MustLoad(fb.LoadFixture("waves/res/getinfo.json", bestBlockHash)),
				ResponseCode: http.StatusOK,
			})

			info, err := client.GetInfo()
			Expect(err).ToNot(HaveOccurred())

			Expect(info).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Chain":        Equal("main"),
					"CurrentBlock": Equal(fmt.Sprintf("%d", bestBlockHash)),
					"BlockHeight":  Equal(bestBlockHash),
				}),
			})))
		})
	})

	Describe("#GetBalance", func() {
		It("Should return the Waves balance transformed to the common output", func() {
			addr := "3PQxNpso2uNbiPM7PQWJMNeYkVsUv4P5mLm"
			balRes := 262056011619

			mockServer.Expect(test.ExpectedCall{
				Path:         "/addresses/balance/details/3PQxNpso2uNbiPM7PQWJMNeYkVsUv4P5mLm",
				Method:       http.MethodGet,
				Response:     MustLoad(fb.LoadFixture("waves/res/getbalance.json", addr, balRes)),
				ResponseCode: http.StatusOK,
			})

			balance, err := client.GetBalance(addr)
			Expect(err).ToNot(HaveOccurred())

			Expect(balance).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Assets": ConsistOf(
						MatchAllFields(Fields{
							"Asset":   Equal("WAVES"),
							"Balance": Equal(fmt.Sprintf("%d", balRes)),
						}),
					),
				}),
			})))
		})
	})

	Describe("#GetTransactionByHash", func() {
		It("Should return the Waves transaction transformed to the common transaction interface", func() {
			txID := "9JnjjmKV5e9h24hKDaGu1tZnFcKLFgQWUzXP9E98UtKc"

			mockServer.Expect(test.ExpectedCall{
				Path:         "/transactions/info/9JnjjmKV5e9h24hKDaGu1tZnFcKLFgQWUzXP9E98UtKc",
				Method:       http.MethodGet,
				Response:     MustLoad(fb.LoadFixture("waves/res/gettransaction.json", txID)),
				ResponseCode: http.StatusOK,
			})

			tx, err := client.GetTransactionByHash(txID)
			Expect(err).ToNot(HaveOccurred())

			Expect(tx).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Transaction": MatchAllFields(Fields{
						"ID":    Equal(txID),
						"From":  Equal("3P8Z5vqm2ECLUc6Dsb1nFQXx84efeSqsv8h"),
						"To":    Equal("3P8wPvtfruNZjpqZACNjdqbtGRphwytdo6D"),
						"Value": Equal("32149900000"),
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
