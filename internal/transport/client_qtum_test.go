package transport_test

import (
	"github.com/hugorut/coins-oracle/pkg/transport"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	. "github.com/hugorut/coins-oracle/internal/transport"
	"github.com/hugorut/coins-oracle/pkg/test"
)

var _ = Describe("QtumClient", func() {
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

		client = &QtumClient{
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
		It("Should return the Qtum node information transformed to the common output", func() {
			bestBlockHash := "hash"
			height := 3445

			mockServer.Expect(test.ExpectedCall{
				Path:   "/api/blocks",
				Method: http.MethodGet,
				QueryParams: map[string]string{
					"date": time.Now().Format("2006-01-02"),
				},
				Response:     MustLoad(fb.LoadFixture("qtum/res/getinfo.json", bestBlockHash, height)),
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
		It("Should return the Qtum balance transformed to the common output", func() {
			addr := "addr"
			balRes := "3444"

			mockServer.Expect(test.ExpectedCall{
				Path:         "/api/address/" + addr,
				Method:       http.MethodGet,
				Response:     MustLoad(fb.LoadFixture("qtum/res/getbalance.json", balRes)),
				ResponseCode: http.StatusOK,
			})

			balance, err := client.GetBalance(addr)
			Expect(err).ToNot(HaveOccurred())

			Expect(balance).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Assets": ConsistOf(
						MatchAllFields(Fields{
							"Asset":   Equal("QTUM"),
							"Balance": Equal(balRes),
						}),
					),
				}),
			})))
		})
	})

	Describe("#GetTransactionByHash", func() {
		It("Should return the Qtum transaction transformed to the common transaction interface", func() {
			txID := "tx-id"

			mockServer.Expect(test.ExpectedCall{
				Path:         "/api/tx/" + txID,
				Method:       http.MethodGet,
				Response:     MustLoad(fb.LoadFixture("qtum/res/gettransaction.json")),
				ResponseCode: http.StatusOK,
			})

			tx, err := client.GetTransactionByHash(txID)
			Expect(err).ToNot(HaveOccurred())

			Expect(tx).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Transaction": MatchAllFields(Fields{
						"ID":    Equal(txID),
						"From":  Equal("QcFkeoah4aD4khsshqdYXcxFn6BTfWGaLY"),
						"To":    Equal("QcWRvSuZx5keXZL8LDrnxyiLkPMYW1sMqj"),
						"Value": Equal("1.464000"),
						"Confirmations": MatchAllFields(Fields{
							"Threshold": PointTo(Equal(int64(5))),
							"Confirmed": BeTrue(),
							"Value":     PointTo(Equal(int64(686))),
						}),
					}),
				}),
			})))
		})
	})
})
