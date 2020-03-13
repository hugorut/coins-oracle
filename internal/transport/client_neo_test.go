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

var _ = Describe("NeoClient", func() {
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

		client = &NeoClient{
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
		It("Should return the Neo node information transformed to the common output", func() {
			bestBlockHash := "0x02a4a4d60250946290aba80fb08b3df9bf81dc590c2e44fbdea47df45aada980"
			count := 1234

			mockServer.Expect(test.ExpectedCall{
				Path:   "/",
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "Application/Json",
				},
				Body:         MustLoad(fb.LoadFixture("neo/req/getinfo.json")),
				Response:     MustLoad(fb.LoadFixture("neo/res/getinfo.json", bestBlockHash)),
				ResponseCode: http.StatusOK,
			}).Then(test.ExpectedCall{
				Path:   "/",
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "Application/Json",
				},
				Body:         MustLoad(fb.LoadFixture("neo/req/getblockcount.json")),
				Response:     MustLoad(fb.LoadFixture("neo/res/getblockcount.json", count)),
				ResponseCode: http.StatusOK,
			})

			info, err := client.GetInfo()
			Expect(err).ToNot(HaveOccurred())

			Expect(info).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Chain":        Equal("main"),
					"CurrentBlock": Equal(bestBlockHash),
					"BlockHeight":  Equal(count),
				}),
			})))
		})
	})

	Describe("#GetBalance", func() {
		It("Should return the Neo balance transformed to the common output", func() {
			addr := "AQVh2pG732YvtNaxEGkQUei3YA4cvo7d2i"
			balRes := "94"

			mockServer.Expect(test.ExpectedCall{
				Path:   "/",
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "Application/Json",
				},
				Body:         MustLoad(fb.LoadFixture("neo/req/getbalance.json", addr)),
				Response:     MustLoad(fb.LoadFixture("neo/res/getbalance.json", balRes)),
				ResponseCode: http.StatusOK,
			})

			balance, err := client.GetBalance(addr)
			Expect(err).ToNot(HaveOccurred())

			Expect(balance).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Assets": ConsistOf(
						MatchAllFields(Fields{
							"Asset":   Equal("c56f33fc6ecfcd0c225c4ab356fee59390af8560be0e930faebe74a6daff7c9b"),
							"Balance": Equal(balRes),
						}),
					),
				}),
			})))
		})
	})

	Describe("#GetTransactionByHash", func() {
		It("Should return the Neo transaction transformed to the common transaction interface", func() {
			txID := "f4250dab094c38d8265acc15c366dc508d2e14bf5699e12d9df26577ed74d657"
			senderID := "abe82713f756eaeebf6fa6440057fca7c36b6c157700738bc34d3634cb765819"

			mockServer.Expect(test.ExpectedCall{
				Path:   "/",
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "Application/Json",
				},
				Body:         MustLoad(fb.LoadFixture("neo/req/gettransaction.json", txID)),
				Response:     MustLoad(fb.LoadFixture("neo/res/gettransaction.json")),
				ResponseCode: http.StatusOK,
			}).Then(test.ExpectedCall{
				Path:   "/",
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "Application/Json",
				},
				Body:         MustLoad(fb.LoadFixture("neo/req/gettransaction.json", senderID)),
				Response:     MustLoad(fb.LoadFixture("neo/res/gettransaction_sender.json")),
				ResponseCode: http.StatusOK,
			})

			tx, err := client.GetTransactionByHash(txID)
			Expect(err).ToNot(HaveOccurred())

			Expect(tx).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Transaction": MatchAllFields(Fields{
						"ID":    Equal(txID),
						"From":  Equal("ALDCagdWUVV4wYoEzCcJ4dtHqtWhsNEEaR"),
						"To":    Equal("AHCNSDkh2Xs66SzmyKGdoDKY752uyeXDrt"),
						"Value": Equal("2950"),
						"Confirmations": MatchAllFields(Fields{
							"Threshold": PointTo(Equal(int64(5))),
							"Confirmed": BeTrue(),
							"Value":     PointTo(Equal(int64(144))),
						}),
					}),
				}),
			})))
		})
	})
})
