package transport_test

import (
	"math/big"
	"net/http"
	"net/url"
	"os"
	"path"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/hugorut/coins-oracle/pkg/transport"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	. "github.com/hugorut/coins-oracle/internal/transport"
	"github.com/hugorut/coins-oracle/pkg/test"
)

var _ = Describe("ERC20Client", func() {
	var (
		fb     *test.FixtureBox
		client transport.CoinClient

		mockServer      *test.Server
		etherScanServer *test.Server
	)

	BeforeEach(func() {
		dir, err := os.Getwd()
		Expect(err).ToNot(HaveOccurred())

		fb = &test.FixtureBox{
			Base: path.Join(dir, "../test/fixtures"),
		}

		mockServer = test.NewTestServer(GinkgoT())
		etherScanServer = test.NewTestServer(GinkgoT())

		ethC, err := ethclient.Dial(mockServer.HttpTest.URL)
		Expect(err).ToNot(HaveOccurred())

		address := common.HexToAddress(ERC20Tokens[TetherAssetID].ContractAddr)
		Expect(err).ToNot(HaveOccurred())

		u, _ := url.Parse(etherScanServer.HttpTest.URL)

		client = &ERC20Client{
			AssetID:      ERC20Tokens[TetherAssetID].AssetID,
			ContractAddr: &address,
			EthClient:    ethC,
			ABIClient: transport.BaseClient{
				BaseURL: u,
				Client:  http.DefaultClient,
			},
			ABIMap: map[string]abi.ABI{},
			MU:     &sync.Mutex{},
		}
	})

	AfterEach(func() {
		mockServer.Close()
		etherScanServer.Close()
	})

	Describe("#GetInfo", func() {
		It("Should return the Tether node information transformed to the common output", func() {
			info, err := client.GetInfo()
			Expect(err).ToNot(HaveOccurred())

			Expect(info).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Chain":        Equal("main"),
					"CurrentBlock": Equal("ERC20"),
					"BlockHeight":  BeZero(),
				}),
			})))
		})
	})

	Describe("#GetBalance", func() {
		It("Should return the Tether balance transformed to the common output", func() {
			addr := "0xd6bd8F134262109E36EA70ee89548B0Bc8bF6D0c"

			contractAddr := "0xdac17f958d2ee523a2206206994597c13d831ec7"
			data := "0x70a08231000000000000000000000000d6bd8f134262109e36ea70ee89548b0bc8bf6d0c"

			balRes := "0x0000000000000000000000000000000000000000000000000000000008fda518"

			mockServer.Expect(test.ExpectedCall{
				Path:   "/",
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "Application/Json",
				},
				Body:         MustLoad(fb.LoadFixture("erc20/req/getbalance.json", contractAddr, data)),
				Response:     MustLoad(fb.LoadFixture("erc20/res/getbalance.json", balRes)),
				ResponseCode: http.StatusOK,
			})

			balance, err := client.GetBalance(addr)
			Expect(err).ToNot(HaveOccurred())

			Expect(balance).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Assets": ConsistOf(
						MatchAllFields(Fields{
							"Asset":   Equal("USDT"),
							"Balance": Equal("150840600"),
						}),
					),
				}),
			})))
		})
	})

	Describe("#GetTransactionByHash", func() {
		It("Should return the Tether transaction transformed to the common transaction interface", func() {
			txID := "0xeeb74ccde78183e6468376f76d7670f1a8eeaa1f13ae2152f7c8afe6b5f51125"
			blockNumber := big.NewInt(8027412)
			currentBlockHeight := big.NewInt(8027418)
			to := "0xdac17f958d2ee523a2206206994597c13d831ec7"

			mockServer.Expect(test.ExpectRPCJsonSuccess(MustLoad(fb.LoadFixture("erc20/req/eth_getBlockByNumber.json", 1)), MustLoad(fb.LoadFixture("erc20/res/eth_getBlockByNumber.json", hexutil.EncodeBig(currentBlockHeight))))).
				Then(test.ExpectRPCJsonSuccess(MustLoad(fb.LoadFixture("erc20/req/eth_getTransactionByHash.json", txID)), MustLoad(fb.LoadFixture("erc20/res/eth_getTransactionByHash.json", txID, to)))).
				Then(test.ExpectRPCJsonSuccess(MustLoad(fb.LoadFixture("erc20/req/eth_chainId.json")), MustLoad(fb.LoadFixture("erc20/res/eth_chainId.json")))).
				Then(test.ExpectRPCJsonSuccess(MustLoad(fb.LoadFixture("erc20/req/eth_getTransactionReceipt.json", txID)), MustLoad(fb.LoadFixture("erc20/res/eth_getTransactionReceipt.json", hexutil.EncodeBig(blockNumber)))))

			etherScanServer.Expect(test.ExpectedCall{
				Path:         "/api",
				Method:       http.MethodGet,
				QueryParams: map[string]string{
					"module":  "contract",
					"action":  "getabi",
					"address": to,
				},
				Response:     MustLoad(fb.LoadFixture("erc20/res/contract_abi.json")),
				ResponseCode: http.StatusOK,
			})
			tx, err := client.GetTransactionByHash(txID)
			Expect(err).ToNot(HaveOccurred())

			Expect(tx).To(PointTo(MatchAllFields(Fields{
				"Data": MatchAllFields(Fields{
					"Transaction": MatchAllFields(Fields{
						"ID":    Equal(txID),
						"From":  Equal("0x59C9cBb043aE0c437676cCfB2c143073c2E2B359"),
						"To":    Equal("0xdAC17F958D2ee523a2206206994597C13D831ec7"),
						"Value": Equal("5000000"),
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
