package handlers_test

import (
	"fmt"
	"github.com/hugorut/coins-oracle/pkg/transport"
	"net/http"
	"net/http/httptest"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/hugorut/coins-oracle/internal/handlers"
	mock_echo "github.com/hugorut/coins-oracle/internal/handlers/mocks"
	mock_transport "github.com/hugorut/coins-oracle/internal/transport/mocks"
)

var _ = Describe("Wallets", func() {
	var (
		e      *echo.Echo
		client *mock_transport.MockCoinClient
		ctrl   *gomock.Controller
		logger *mock_echo.MockLogger
	)

	BeforeEach(func() {
		e = echo.New()
		ctrl = gomock.NewController(GinkgoT())
		client = mock_transport.NewMockCoinClient(ctrl)
		logger = mock_echo.NewMockLogger(ctrl)

		logger.EXPECT().Print(gomock.Any()).AnyTimes()
		logger.EXPECT().Printf(gomock.Any()).AnyTimes()

		e.Logger = logger
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("GetTransactionByHash", func() {
		It("Should render a json response", func() {
			assetID := "test-node"
			hash := "hash1234"

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/nodes/%s/txs/%s", assetID, hash), nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			c.SetParamNames("assetId", "txHash")
			c.SetParamValues(assetID, hash)

			c.Set("coin_client", client)

			var threshold int64 = 5
			var confirmationsValue int64 = 1

			client.EXPECT().GetTransactionByHash(gomock.Eq(hash)).Return(&transport.TransactionResp{
				Data: struct {
					Transaction transport.Transaction `json:"transaction"`
				}{
					Transaction: transport.Transaction{
						ID:    "0x88df016429689c079f3b2f6ad39fa052532c56795b733da78a91ebe6a713944b",
						From:  "0xa7d9ddbe1f17865597fbd27ec712455208b6b76d",
						To:    "0xf02c1c8e6114b1dbe8937a39260b5b0a374432bb",
						Value: "4290000000000000",
						Confirmations: transport.Confirmations{
							Threshold: &threshold,
							Confirmed: false,
							Value:     &confirmationsValue,
						},
					},
				},
			}, nil)

			err := GetTransactionByHash(c)
			Expect(err).ToNot(HaveOccurred())

			Expect(rec.Body.String()).Should(MatchJSON(`{
					"data": {
						"transaction": {
							"id":"0x88df016429689c079f3b2f6ad39fa052532c56795b733da78a91ebe6a713944b",
							"from":"0xa7d9ddbe1f17865597fbd27ec712455208b6b76d",
							"to":"0xf02c1c8e6114b1dbe8937a39260b5b0a374432bb",
							"value":"4290000000000000",
							"confirmations": {
								"threshold": 5,
								"confirmed": false,
								"value": 1
							}
					  }
					}
				}`))
			Expect(rec.Code).To(Equal(http.StatusOK))
		})
	})
})
