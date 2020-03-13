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

var _ = Describe("NanoClient", func() {
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

		client = &NanoClient{
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
		It("Should return the Nano node information transformed to the common output", func() {
			bestBlockHash := "32281218"

			mockServer.Expect(test.ExpectedCall{
				Path:   "/",
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "Application/Json",
				},
				Body:         MustLoad(fb.LoadFixture("nano/req/getinfo.json")),
				Response:     MustLoad(fb.LoadFixture("nano/res/getinfo.json", bestBlockHash)),
				ResponseCode: http.StatusOK,
			})

			info, err := client.GetInfo()
			Expect(err).ToNot(HaveOccurred())

			Expect(info).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Chain":        Equal("main"),
					"CurrentBlock": Equal(bestBlockHash),
					"BlockHeight":  Equal(32281218),
				}),
			})))
		})
	})

	Describe("#GetBalance", func() {
		It("Should return the Nano balance transformed to the common output", func() {
			addr := "xrb_3t6k35gi95xu6tergt6p69ck76ogmitsa8mnijtpxm9fkcm736xtoncuohr3"
			balRes := "325586539664609129644855132177"

			mockServer.Expect(test.ExpectedCall{
				Path:   "/",
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "Application/Json",
				},
				Body:         MustLoad(fb.LoadFixture("nano/req/getbalance.json", addr)),
				Response:     MustLoad(fb.LoadFixture("nano/res/getbalance.json", balRes)),
				ResponseCode: http.StatusOK,
			})

			balance, err := client.GetBalance(addr)
			Expect(err).ToNot(HaveOccurred())

			Expect(balance).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Assets": ConsistOf(
						MatchAllFields(Fields{
							"Asset":   Equal("NANO"),
							"Balance": Equal("325586539664609129644855132177"),
						}),
					),
				}),
			})))
		})
	})

	Describe("#GetTransactionByHash", func() {
		It("Should return the Nano transaction transformed to the common transaction interface", func() {
			txID := "87434F8041869A01C8F6F263B87972D7BA443A72E0A97D7A3FD0CCC2358FD6F9"

			mockServer.Expect(test.ExpectedCall{
				Path:   "/",
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "Application/Json",
				},
				Body:         MustLoad(fb.LoadFixture("nano/req/gettransaction.json", txID)),
				Response:     MustLoad(fb.LoadFixture("nano/res/gettransaction.json")),
				ResponseCode: http.StatusOK,
			})

			tx, err := client.GetTransactionByHash(txID)
			Expect(err).ToNot(HaveOccurred())

			Expect(tx).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Transaction": MatchAllFields(Fields{
						"ID":    Equal(txID),
						"From":  Equal("xrb_3q5rs69fjjgci9bn9mwair713dad5cdjpaonmr6odxooccgnnw7xiej85o9c"),
						"To":    Equal("xrb_19f9kzibbshoc94k3cz4iw5e6d4c57r5i8tka96e8fnzc6wnyrk5mrj91m9z"),
						"Value": Equal("100000000000000000000000000000"),
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
