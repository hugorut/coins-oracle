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

var _ = Describe("TronClient", func() {
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

		client = &TronClient{
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
		It("Should return the Tron balance transformed to the common output", func() {
			addr := "TWsm8HtU2A5eEzoT8ev8yaoFjHsXLLrckb"
			hexAddr := "41E552F6487585C2B58BC2C9BB4492BC1F17132CD0"
			balRes := 4710382

			mockServer.Expect(test.ExpectedCall{
				Path:   "/wallet/getaccount",
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "Application/Json",
				},
				Body:         MustLoad(fb.LoadFixture("tron/req/getbalance.json", hexAddr)),
				Response:     MustLoad(fb.LoadFixture("tron/res/getbalance.json", hexAddr, balRes)),
				ResponseCode: http.StatusOK,
			})

			balance, err := client.GetBalance(addr)
			Expect(err).ToNot(HaveOccurred())

			Expect(balance).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Assets": ConsistOf(
						MatchAllFields(Fields{
							"Asset":   Equal("TRX"),
							"Balance": Equal(fmt.Sprintf("%d", balRes)),
						}),
					),
				}),
			})))
		})
	})

	Describe("#GetInfo", func() {
		It("Should return the Tron node information transformed to the common output", func() {
			bestBlockHash := "0000000000006a5011fe7c20bf354549138002e77f1035d6b301dc20757ba8c4"

			mockServer.Expect(test.ExpectedCall{
				Path:   "/wallet/getnowblock",
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "Application/Json",
				},
				Response:     MustLoad(fb.LoadFixture("tron/res/getinfo.json", bestBlockHash)),
				ResponseCode: http.StatusOK,
			})

			info, err := client.GetInfo()
			Expect(err).ToNot(HaveOccurred())

			Expect(info).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Chain":        Equal("main"),
					"CurrentBlock": Equal(bestBlockHash),
					"BlockHeight":  Equal(27216),
				}),
			})))
		})
	})

	Describe("#GetTransactionByHash", func() {
		It("Should return the Tron transaction transformed to the common transaction interface", func() {
			txID := "0ab16c340e85d52c1179aceca5133e711a5850f4ea42d5bac2e8929a3330551d"
			bestBlockHash := "0000000000006a5011fe7c20bf354549138002e77f1035d6b301dc20757ba8c4"

			mockServer.Expect(test.ExpectedCall{
				Path:   "/wallet/gettransactionbyid",
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "Application/Json",
				},
				Body:         MustLoad(fb.LoadFixture("tron/req/gettransaction.json", txID)),
				Response:     MustLoad(fb.LoadFixture("tron/res/gettransaction.json", txID)),
				ResponseCode: http.StatusOK,
			}).Then(test.ExpectedCall{
				Path:   "/wallet/gettransactioninfobyid",
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "Application/Json",
				},
				Body:         MustLoad(fb.LoadFixture("tron/req/gettransactioninfo.json", txID)),
				Response:     MustLoad(fb.LoadFixture("tron/res/gettransactioninfo.json", txID)),
				ResponseCode: http.StatusOK,
			}).Then(test.ExpectedCall{
				Path:   "/wallet/getnowblock",
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "Application/Json",
				},
				Response:     MustLoad(fb.LoadFixture("tron/res/getinfo.json", bestBlockHash)),
				ResponseCode: http.StatusOK,
			})

			tx, err := client.GetTransactionByHash(txID)
			Expect(err).ToNot(HaveOccurred())

			Expect(tx).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Transaction": MatchAllFields(Fields{
						"ID":    Equal(txID),
						"From":  Equal("TNDFkUNA2TukukC1Moeqj61pAS53NFchGF"),
						"To":    Equal("TTC9XSNGgftzXxFbtc3VfSAepMMAL6RfiD"),
						"Value": Equal("37"),
						"Confirmations": MatchAllFields(Fields{
							"Threshold": PointTo(Equal(int64(5))),
							"Confirmed": BeTrue(),
							"Value":     PointTo(Equal(int64(19))),
						}),
					}),
				}),
			})))
		})
	})
})
