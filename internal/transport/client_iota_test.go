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

var _ = Describe("IotaClient", func() {
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

		client = &IotaClient{
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
		It("Should return the Iota node information transformed to the common output", func() {
			bestBlockHash := "hash"
			height := 12323

			mockServer.Expect(test.ExpectedCall{
				Path:         "/live/history",
				Method:       http.MethodGet,
				Response:     MustLoad(fb.LoadFixture("iota/res/getinfo.json", height, bestBlockHash)),
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
		It("Should return the Iota balance transformed to the common output", func() {
			addr := "addr"
			balRes := 6662

			mockServer.Expect(test.ExpectedCall{
				Path:         "/addresses/" + addr,
				Method:       http.MethodGet,
				Response:     MustLoad(fb.LoadFixture("iota/res/getbalance.json", balRes)),
				ResponseCode: http.StatusOK,
			})

			balance, err := client.GetBalance(addr)
			Expect(err).ToNot(HaveOccurred())

			Expect(balance).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Assets": ConsistOf(
						MatchAllFields(Fields{
							"Asset":   Equal("MIOTA"),
							"Balance": Equal("6662"),
						}),
					),
				}),
			})))
		})
	})

	Describe("#GetTransactionByHash", func() {
		It("Should return the Iota transaction transformed to the common transaction interface", func() {
			txID := "transaction-id"
			bundleID := "bundle"

			mockServer.Expect(test.ExpectedCall{
				Path:         "/transactions/" + txID,
				Method:       http.MethodGet,
				Response:     MustLoad(fb.LoadFixture("iota/res/gettransaction.json", bundleID)),
				ResponseCode: http.StatusOK,
			}).Then(test.ExpectedCall{
				Path:         "/bundles/" + bundleID,
				Method:       http.MethodGet,
				Response:     MustLoad(fb.LoadFixture("iota/res/getbundle.json")),
				ResponseCode: http.StatusOK,
			})

			tx, err := client.GetTransactionByHash(txID)
			Expect(err).ToNot(HaveOccurred())

			Expect(tx).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Transaction": MatchAllFields(Fields{
						"ID":    Equal(txID),
						"From":  Equal("TEMCMWCIRPSNLDFNJVY9OPNLAS9YKXIOTFCJFKOMKYWNYPKWEMQKQGPIHWOQKFD9GPXNVNI9C9FJROVYX"),
						"To":    Equal("9QRMOUYLTAFQCOPUCIMYU9MMZNBSPWDLTCGNBATR9YGDYWIIWRLALQMWKUQNHRAFFSZOYNLDSTR9CAUBZ"),
						"Value": Equal("55"),
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
