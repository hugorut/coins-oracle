package transport_test

import (
	"math/big"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	. "github.com/hugorut/coins-oracle/internal/transport"
	"github.com/hugorut/coins-oracle/pkg/test"
)

var _ = Describe("EthereumClassicClient", func() {
	var (
		fb                     *test.FixtureBox
		testKey, _             = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
		testAddr               = crypto.PubkeyToAddress(testKey.PublicKey)
		testBalance            = big.NewInt(2e10)
		fixtureTransactionHash = "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"
		resultingBlockHash     = "0xc6ef2fc5426d6ad6fd9e2a26abeab0aa2411b7ab17f30a99d3cb96aed1d1055b"
	)

	BeforeEach(func() {
		dir, err := os.Getwd()
		Expect(err).ToNot(HaveOccurred())

		fb = &test.FixtureBox{
			Base: path.Join(dir, "../test/fixtures"),
		}
	})

	Describe("#GetBalance", func() {
		It("Should conform rpc output to standard balance rep", func() {
			reqBody, err := fb.LoadFixture("ethereum/req/eth_getBalance.json", strings.ToLower(testAddr.String()))
			Expect(err).ToNot(HaveOccurred())

			resBody, err := fb.LoadFixture("ethereum/res/eth_getBalance.json", (*hexutil.Big)(testBalance).String())
			Expect(err).ToNot(HaveOccurred())

			server := test.NewTestServer(GinkgoT(), test.ExpectedCall{
				Path:   "/",
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "Application/Json",
				},
				Body:         reqBody,
				Response:     resBody,
				ResponseCode: http.StatusOK,
			})
			defer server.Close()

			client, err := ethclient.Dial(server.HttpTest.URL)
			Expect(err).ToNot(HaveOccurred())

			ec := EthereumClassicClient{
				EthereumClient: &EthereumClient{AssetID: EthereumclassicAssetID, Client: client},
			}

			b, err := ec.GetBalance(testAddr.String())
			Expect(err).ToNot(HaveOccurred())

			Expect(b).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Assets": ConsistOf(
						MatchAllFields(Fields{
							"Asset":   Equal("ETC"),
							"Balance": Equal(testBalance.String()),
						}),
					),
				}),
			})))
		})
	})

	Describe("#GetInfo", func() {
		It("Should conform rpc output to standard balance rep", func() {
			server := test.NewTestServer(GinkgoT(), test.ExpectedCall{
				Path:   "/",
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "Application/Json",
				},
				Body:         MustLoad(fb.LoadFixture("ethereum/req/net_version.json")),
				Response:     MustLoad(fb.LoadFixture("ethereum/res/net_version.json")),
				ResponseCode: http.StatusOK,
			},
				test.ExpectedCall{
					Path:   "/",
					Method: "POST",
					Headers: map[string]string{
						"Content-Type": "Application/Json",
					},
					Body:         MustLoad(fb.LoadFixture("ethereum/req/eth_getBlockByNumber.json", 2)),
					Response:     MustLoad(fb.LoadFixture("ethereum/res/eth_getBlockByNumber.json")),
					ResponseCode: http.StatusOK,
				},
			)
			defer server.Close()

			client, err := ethclient.Dial(server.HttpTest.URL)
			Expect(err).ToNot(HaveOccurred())

			ec := EthereumClient{
				Client: client,
			}

			b, err := ec.GetInfo()
			Expect(err).ToNot(HaveOccurred())

			Expect(b).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Chain":        Equal("3"),
					"CurrentBlock": Equal(fixtureTransactionHash),
					"BlockHeight":  Equal(15061313),
				}),
			})))
		})
	})

	Describe("#GetTransactionByHash", func() {
		It("Should return a transaction in standardised format", func() {
			server := test.NewTestServer(
				GinkgoT(),
				test.ExpectRPCJsonSuccess(MustLoad(fb.LoadFixture("ethereum/req/eth_getBlockByNumber.json", 1)), MustLoad(fb.LoadFixture("ethereum/res/eth_getBlockByNumber.json"))),
				test.ExpectRPCJsonSuccess(MustLoad(fb.LoadFixture("ethereum/req/eth_getTransactionByHash.json", fixtureTransactionHash)), MustLoad(fb.LoadFixture("ethereum/res/eth_getTransactionByHash.json", fixtureTransactionHash))),
				test.ExpectRPCJsonSuccess(MustLoad(fb.LoadFixture("ethereum/req/eth_chainId.json")), MustLoad(fb.LoadFixture("ethereum/res/eth_chainId.json"))),
				test.ExpectRPCJsonSuccess(MustLoad(fb.LoadFixture("ethereum/req/eth_getTransactionReceipt.json", "0x88df016429689c079f3b2f6ad39fa052532c56795b733da78a91ebe6a713944b")), MustLoad(fb.LoadFixture("ethereum/res/eth_getTransactionReceipt.json", "0x88df016429689c079f3b2f6ad39fa052532c56795b733da78a91ebe6a713944b", resultingBlockHash))),
			)
			defer server.Close()

			client, err := ethclient.Dial(server.HttpTest.URL)
			Expect(err).ToNot(HaveOccurred())

			ec := EthereumClient{
				Client: client,
			}

			tran, err := ec.GetTransactionByHash(fixtureTransactionHash)
			Expect(err).ToNot(HaveOccurred())

			Expect(tran).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Transaction": MatchAllFields(Fields{
						"ID":    Equal(fixtureTransactionHash),
						"From":  Equal("0xa7d9ddBE1f17865597fBD27EC712455208B6B76d"),
						"To":    Equal("0xF02c1c8e6114b1Dbe8937a39260b5b0a374432bB"),
						"Value": Equal("4290000000000000"),
						"Confirmations": MatchAllFields(Fields{
							"Threshold": PointTo(Equal(int64(5))),
							"Confirmed": BeTrue(),
							"Value":     PointTo(Equal(int64(15061302))),
						}),
					}),
				}),
			})))
		})
	})
})
