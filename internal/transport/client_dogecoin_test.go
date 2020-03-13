package transport_test

import (
	"github.com/hugorut/coins-oracle/pkg/transport"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/btcsuite/btcd/rpcclient"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	. "github.com/hugorut/coins-oracle/internal/transport"
	"github.com/hugorut/coins-oracle/pkg/test"
)

var _ = Describe("DogecoinClient", func() {
	var (
		fb     *test.FixtureBox
		client transport.CoinClient

		mockServer *test.Server

		rpcUser string = "user"
		rpcPass string = "pass"
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

		connCfg := &rpcclient.ConnConfig{
			Host:         u.Host,
			User:         rpcUser,
			Pass:         rpcPass,
			HTTPPostMode: true,
			DisableTLS:   true,
		}

		btcClient, err := rpcclient.New(connCfg, nil)
		Expect(err).ToNot(HaveOccurred())

		client = DogecoinClient{
			BitcoinClient: &BitcoinClient{
				AssetID: DogecoinAssetID,
				Client:  btcClient,
			},
		}
	})

	AfterEach(func() {
		mockServer.Close()
	})

	Describe("#GetBalance", func() {
		It("Should return the Bitcoin balance transformed to the common output", func() {
			addr := "3EdTTxcfptcBziNR1YH3pdcWdQ923jSXaR"

			mockServer.Expect(test.ExpectedCall{
				Path:   "/",
				Method: "POST",
				Headers: map[string]string{
					"Content-Type":  "Application/Json",
					"Authorization": "Basic " + test.BasicAuth(rpcUser, rpcPass),
				},
				Body:         MustLoad(fb.LoadFixture("bitcoin/req/listunspent.json", addr)),
				Response:     MustLoad(fb.LoadFixture("bitcoin/res/listunspent.json", 0.00462265)),
				ResponseCode: http.StatusOK,
			})

			balance, err := client.GetBalance(addr)
			Expect(err).ToNot(HaveOccurred())

			Expect(balance).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Assets": ConsistOf(
						MatchAllFields(Fields{
							"Asset":   Equal("DOGE"),
							"Balance": Equal("0.004623"),
						}),
					),
				}),
			})))
		})
	})

	Describe("#GetInfo", func() {
		It("Should return the Bitcoin node information transformed to the common output", func() {
			bestBlockHash := "current-block-123"

			mockServer.Expect(test.ExpectedCall{
				Path:   "/",
				Method: "POST",
				Headers: map[string]string{
					"Content-Type":  "Application/Json",
					"Authorization": "Basic " + test.BasicAuth(rpcUser, rpcPass),
				},
				Body:         MustLoad(fb.LoadFixture("bitcoin/req/getinfo.json")),
				Response:     MustLoad(fb.LoadFixture("bitcoin/res/getinfo.json", bestBlockHash)),
				ResponseCode: http.StatusOK,
			})

			info, err := client.GetInfo()
			Expect(err).ToNot(HaveOccurred())

			Expect(info).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Chain":        Equal("main"),
					"CurrentBlock": Equal(bestBlockHash),
					"BlockHeight":  Equal(595303),
				}),
			})))
		})
	})

	Describe("#GetTransactionByHash", func() {
		It("Should return the Bitcoin transaction transformed to the common transaction interface", func() {
			txID := "6f5dfa31bef79d0c8cdd58530fc9f0ed2427e7085d421755f3fe78ca6ac326ef"
			senderID := "c88f369cfe24e402eafd97c7318183ed780baa3b92a3459fc161eb9472ea532b"

			mockServer.Expect(test.ExpectedCall{
				Path:   "/",
				Method: "POST",
				Headers: map[string]string{
					"Content-Type":  "Application/Json",
					"Authorization": "Basic " + test.BasicAuth(rpcUser, rpcPass),
				},
				Body:         MustLoad(fb.LoadFixture("bitcoin/req/getrawtransaction.json", 1, txID)),
				Response:     MustLoad(fb.LoadFixture("bitcoin/res/getrawtransaction.json")),
				ResponseCode: http.StatusOK,
			}).Then(test.ExpectedCall{
				Path:   "/",
				Method: "POST",
				Headers: map[string]string{
					"Content-Type":  "Application/Json",
					"Authorization": "Basic " + test.BasicAuth(rpcUser, rpcPass),
				},
				Body:         MustLoad(fb.LoadFixture("bitcoin/req/decoderawtransaction.json")),
				Response:     MustLoad(fb.LoadFixture("bitcoin/res/decoderawtransaction.json", txID, txID)),
				ResponseCode: http.StatusOK,
			}).Then(test.ExpectedCall{
				Path:   "/",
				Method: "POST",
				Headers: map[string]string{
					"Content-Type":  "Application/Json",
					"Authorization": "Basic " + test.BasicAuth(rpcUser, rpcPass),
				},
				Body:         MustLoad(fb.LoadFixture("bitcoin/req/getrawtransaction.json", 3, senderID)),
				Response:     MustLoad(fb.LoadFixture("bitcoin/res/getrawtransaction_sender.json")),
				ResponseCode: http.StatusOK,
			}).Then(test.ExpectedCall{
				Path:   "/",
				Method: "POST",
				Headers: map[string]string{
					"Content-Type":  "Application/Json",
					"Authorization": "Basic " + test.BasicAuth(rpcUser, rpcPass),
				},
				Body:         MustLoad(fb.LoadFixture("bitcoin/req/decoderawtransaction_sender.json")),
				Response:     MustLoad(fb.LoadFixture("bitcoin/res/decoderawtransaction_sender.json", senderID, senderID)),
				ResponseCode: http.StatusOK,
			})

			tx, err := client.GetTransactionByHash(txID)
			Expect(err).ToNot(HaveOccurred())

			Expect(tx).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Transaction": MatchAllFields(Fields{
						"ID":    Equal(txID),
						"From":  Equal("1NULeRwToV4pm394qnmoK6F2Gyu6DGUwq4"),
						"To":    Equal("1Hb1xsuhehKYcvkTRjWUxkF4Lh75kifZZh"),
						"Value": Equal("25.098818"),
						"Confirmations": MatchAllFields(Fields{
							"Threshold": PointTo(Equal(int64(5))),
							"Confirmed": BeTrue(),
							"Value":     PointTo(Equal(int64(46413))),
						}),
					}),
				}),
			})))
		})
	})
})
