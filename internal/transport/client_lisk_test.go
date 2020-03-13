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

var _ = Describe("LiskClient", func() {
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

		client = &LiskClient{
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
		It("Should return the Lisk node information transformed to the common output", func() {
			bestBlockHash := "12310941572803966337"

			mockServer.Expect(test.ExpectedCall{
				Path: "/api/blocks",
				QueryParams: map[string]string{
					"limit": "1",
					"sort":  "height:desc",
				},
				Method:       http.MethodGet,
				Response:     MustLoad(fb.LoadFixture("lisk/res/getinfo.json", bestBlockHash)),
				ResponseCode: http.StatusOK,
			})

			info, err := client.GetInfo()
			Expect(err).ToNot(HaveOccurred())

			Expect(info).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Chain":        Equal("main"),
					"CurrentBlock": Equal(bestBlockHash),
					"BlockHeight":  Equal(10406987),
				}),
			})))
		})
	})

	Describe("#GetBalance", func() {
		It("Should return the Lisk balance transformed to the common output", func() {
			addr := "7714731151444318219L"
			balRes := "19905973045"

			mockServer.Expect(test.ExpectedCall{
				Path:   "/api/accounts",
				Method: http.MethodGet,
				QueryParams: map[string]string{
					"address": addr,
					"limit":   "1",
				},
				Response:     MustLoad(fb.LoadFixture("lisk/res/getbalance.json", addr, balRes)),
				ResponseCode: http.StatusOK,
			})

			balance, err := client.GetBalance(addr)
			Expect(err).ToNot(HaveOccurred())

			Expect(balance).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Assets": ConsistOf(
						MatchAllFields(Fields{
							"Asset":   Equal("LSK"),
							"Balance": Equal(balRes),
						}),
					),
				}),
			})))
		})
	})

	Describe("#GetTransactionByHash", func() {
		It("Should return the Lisk transaction transformed to the common transaction interface", func() {
			txID := "6980013695783136273"

			mockServer.Expect(test.ExpectedCall{
				Path:   "/api/transactions",
				Method: http.MethodGet,
				QueryParams: map[string]string{
					"id":    txID,
					"limit": "1",
				},
				Response:     MustLoad(fb.LoadFixture("lisk/res/gettransaction.json")),
				ResponseCode: http.StatusOK,
			})

			tx, err := client.GetTransactionByHash(txID)
			Expect(err).ToNot(HaveOccurred())

			Expect(tx).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Transaction": MatchAllFields(Fields{
						"ID":    Equal(txID),
						"From":  Equal("7714731151444318219L"),
						"To":    Equal("1186872597084592226L"),
						"Value": Equal("33300000000"),
						"Confirmations": MatchAllFields(Fields{
							"Threshold": PointTo(Equal(int64(5))),
							"Confirmed": BeTrue(),
							"Value":     PointTo(Equal(int64(205))),
						}),
					}),
				}),
			})))
		})
	})
})
