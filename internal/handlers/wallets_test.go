package handlers_test

import (
	"errors"
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

	Describe("GetBalance", func() {
		Context("With a successful client output", func() {
			It("Should render a json response", func() {
				assetID := "test-node"
				addr := "address"

				req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/nodes/%s/addrs/%s/balance", assetID, addr), nil)
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
				rec := httptest.NewRecorder()

				c := e.NewContext(req, rec)
				c.SetParamNames("assetId", "addr")
				c.SetParamValues(assetID, addr)

				c.Set("coin_client", client)

				client.EXPECT().GetBalance(gomock.Eq(addr)).Return(&transport.Balance{
					Data: transport.BalanceData{
						Assets: []transport.Asset{
							{
								Asset:   assetID,
								Balance: "14",
							},
						},
					},
				}, nil)

				err := GetWalletBalance(c)
				Expect(err).ToNot(HaveOccurred())

				Expect(rec.Body.String()).Should(MatchJSON(`{
					"data": {
						"assets": [
							{
								"asset": "test-node",
								"balance": "14"
							}
						]
					}
				}`))
				Expect(rec.Code).To(Equal(http.StatusOK))
			})
		})

		Context("With an errored client output", func() {
			It("Should log the error and return", func() {
				assetID := "test-node"
				addr := "address"

				req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/nodes/%s/addrs/%s/balance", assetID, addr), nil)
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
				rec := httptest.NewRecorder()

				c := e.NewContext(req, rec)

				c.SetParamNames("assetId", "addr")
				c.SetParamValues(assetID, addr)

				c.Set("coin_client", client)

				eMess := "test message"
				clientE := errors.New(eMess)
				client.EXPECT().GetBalance(gomock.Any()).Return(&transport.Balance{}, clientE)
				logger.EXPECT().Errorf(gomock.AssignableToTypeOf(""), gomock.Eq(addr), gomock.Eq(assetID), gomock.Eq(clientE))

				err := GetWalletBalance(c)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
