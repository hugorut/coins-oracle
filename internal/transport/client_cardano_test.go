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

var _ = Describe("CardanoClient", func() {
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

		client = &CardanoClient{
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
		It("Should return the Cardano balance transformed to the common output", func() {
			addr := "addr"
			balRes := "7748527449475"

			mockServer.Expect(test.ExpectedCall{
				Path:         "/api/addresses/summary/addr",
				Method:       "GET",
				Response:     MustLoad(fb.LoadFixture("cardano/res/getbalance.json", balRes)),
				ResponseCode: http.StatusOK,
			})

			balance, err := client.GetBalance(addr)
			Expect(err).ToNot(HaveOccurred())

			Expect(balance).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Assets": ConsistOf(
						MatchAllFields(Fields{
							"Asset":   Equal("ADA"),
							"Balance": Equal(balRes),
						}),
					),
				}),
			})))
		})
	})

	Describe("#GetInfo", func() {
		It("Should return the Cardano node information transformed to the common output", func() {
			bestBlockHash := "515606761e5f9a3f2bc5def811f3ff875e151adbe34270247e9412688cc17a81"

			mockServer.Expect(test.ExpectedCall{
				Path:         "/api/blocks/pages",
				Method:       "GET",
				Response:     MustLoad(fb.LoadFixture("cardano/res/getinfo.json", bestBlockHash)),
				ResponseCode: http.StatusOK,
			})

			info, err := client.GetInfo()
			Expect(err).ToNot(HaveOccurred())

			Expect(info).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Chain":        Equal("main"),
					"CurrentBlock": Equal(bestBlockHash),
					"BlockHeight":  Equal(5759),
				}),
			})))
		})
	})

	Describe("#GetTransactionByHash", func() {
		It("Should return the Cardano transaction transformed to the common transaction interface", func() {
			txID := "dfcdf709c046fd85ca373434fc6386f1a27c0d6792cede2b9cf38c6f4e7394b4"

			mockServer.Expect(test.ExpectedCall{
				Path:         "/api/txs/summary/" + txID,
				Method:       "GET",
				Response:     MustLoad(fb.LoadFixture("cardano/res/gettransaction.json", txID)),
				ResponseCode: http.StatusOK,
			})

			tx, err := client.GetTransactionByHash(txID)
			Expect(err).ToNot(HaveOccurred())

			Expect(tx).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Transaction": MatchAllFields(Fields{
						"ID":    Equal(txID),
						"From":  Equal("DdzFFzCqrhsetDGW9EV3pQgGPxeHCyeo7LSruMcvu9U8zUqr23eZVZPL5K6KareJDtQpago7y7R4M3Gd941FvC4BXPULaD8Z8myRiFjj"),
						"To":    Equal("DdzFFzCqrht66tunNTdhEUfKFGE5sAqeJjvafxSJ1u7XEh2GBkAR6SZsGetCorjsR2mRreU7SqqddaEeQ13CtFmiBqonPtMUoHvPWkrX"),
						"Value": Equal("7748527449475"),
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
