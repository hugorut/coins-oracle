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

var _ = Describe("NemClient", func() {
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
		client = &NemClient{
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
		It("Should return the Nem node information transformed to the common output", func() {
			bestBlockHash := "4439eebb0f32a20e18f3bcb4632fd6de6b7ee2207c39ed961b7cabed06d5635d"

			mockServer.Expect(test.ExpectedCall{
				Path:         "/chain/last-block",
				Method:       http.MethodGet,
				Response:     MustLoad(fb.LoadFixture("nem/res/getinfo.json", bestBlockHash)),
				ResponseCode: http.StatusOK,
			})

			info, err := client.GetInfo()
			Expect(err).ToNot(HaveOccurred())

			Expect(info).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Chain":        Equal("main"),
					"CurrentBlock": Equal(bestBlockHash),
					"BlockHeight":  Equal(2355047),
				}),
			})))
		})
	})

	Describe("#GetBalance", func() {
		It("Should return the Nem balance transformed to the common output", func() {
			addr := "NCKZD7JGDLNDIVVPH6U2PG2QKD3PX3FX4CPZMF2A"
			balRes := 8641300704

			mockServer.Expect(test.ExpectedCall{
				Path: "/account/get",
				QueryParams: map[string]string{
					"address": addr,
				},
				Method:       "GET",
				Response:     MustLoad(fb.LoadFixture("nem/res/getbalance.json", addr, balRes)),
				ResponseCode: http.StatusOK,
			})

			balance, err := client.GetBalance(addr)
			Expect(err).ToNot(HaveOccurred())

			Expect(balance).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Assets": ConsistOf(
						MatchAllFields(Fields{
							"Asset":   Equal("XEM"),
							"Balance": Equal(fmt.Sprintf("%d", balRes)),
						}),
					),
				}),
			})))
		})
	})

	Describe("#GetTransactionByHash", func() {
		It("Should return the Nem transaction transformed to the common transaction interface", func() {
			txID := "a4b667fdcd9a4d7e7bef0bfb8d1e488015fb51e7b8f3f6060b1d00313e90da08"
			addr := "from"

			mockServer.Expect(test.ExpectedCall{
				Path:   "/transaction/get",
				Method: "GET",
				QueryParams: map[string]string{
					"hash": txID,
				},
				Response:     MustLoad(fb.LoadFixture("nem/res/gettransaction.json", txID)),
				ResponseCode: http.StatusOK,
			}).Then(test.ExpectedCall{
				Path:   "/account/get/from-public-key",
				Method: "GET",
				QueryParams: map[string]string{
					"publicKey": "d22b047b670fd32ad9fa421a37415ab0677a786b916cad34820bcd395066cd49",
				},
				Response:     MustLoad(fb.LoadFixture("nem/res/getbalance.json", addr, 0)),
				ResponseCode: http.StatusOK,
			}).Then(test.ExpectedCall{
				Path:         "/chain/last-block",
				Method:       http.MethodGet,
				Response:     MustLoad(fb.LoadFixture("nem/res/getinfo.json", "")),
				ResponseCode: http.StatusOK,
			})

			tx, err := client.GetTransactionByHash(txID)
			Expect(err).ToNot(HaveOccurred())

			Expect(tx).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Transaction": MatchAllFields(Fields{
						"ID":    Equal(txID),
						"From":  Equal(addr),
						"To":    Equal("NDWBJQTYMGDV44YHR3RC4BEH5PY75JQVAWSB6MNQ"),
						"Value": Equal("50000"),
						"Confirmations": MatchAllFields(Fields{
							"Threshold": PointTo(Equal(int64(5))),
							"Confirmed": BeTrue(),
							"Value":     PointTo(Equal(int64(6))),
						}),
					}),
				}),
			})))
		})
	})
})
