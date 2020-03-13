package handlers_test

import (
	transport2 "github.com/hugorut/coins-oracle/pkg/transport"
	"net/http"
	"net/http/httptest"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/hugorut/coins-oracle/internal/handlers"
	mock_echo "github.com/hugorut/coins-oracle/internal/handlers/mocks"
	"github.com/hugorut/coins-oracle/internal/transport"
	mock_transport "github.com/hugorut/coins-oracle/internal/transport/mocks"
)

var _ = Describe("Nodes", func() {
	var (
		e      *echo.Echo
		ctrl   *gomock.Controller
		router *mock_transport.MockRouter
		logger *mock_echo.MockLogger
	)

	BeforeEach(func() {
		e = echo.New()
		ctrl = gomock.NewController(GinkgoT())
		logger = mock_echo.NewMockLogger(ctrl)
		router = mock_transport.NewMockRouter(ctrl)

		logger.EXPECT().Print().AnyTimes()
		logger.EXPECT().Printf(gomock.Any()).AnyTimes()
		e.Logger = logger
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("GetNodes", func() {
		It("Should return a list of available nodes", func() {
			req := httptest.NewRequest(http.MethodGet, "/nodes", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)

			c.Set("coin_router", router)

			node1 := transport.CoinNode{
				AssetId: "test1",
				Running: true,
				Info: &transport2.CoinData{
					Chain:        "main",
					BlockHeight:  23,
					CurrentBlock: "15",
				},
			}
			node2 := transport.CoinNode{
				AssetId: "test2",
				Running: false,
				Info: &transport2.CoinData{
					Chain:        "main",
					BlockHeight:  24,
					CurrentBlock: "22",
				},
			}

			router.EXPECT().GetNodes(gomock.Eq(true)).Return([]transport.CoinNode{
				node1,
				node2,
			})

			err := handlers.GetNodes(c)
			Expect(err).ToNot(HaveOccurred())

			Expect(rec.Body.String()).Should(MatchJSON(`{
					"data": {
						"nodes": [
							{
								"assetId": "test1",
  								"running": true,
								"info": {
									"chain": "main",
									"block_height": 23,
									"current_block_hash": "15"
								}
							},
							{
								"assetId": "test2",
  								"running": false,
								"info": {
									"chain": "main",
									"block_height": 24,
									"current_block_hash": "22"
								}
							}
						]
					}
				}`))
			Expect(rec.Code).To(Equal(http.StatusOK))
		})
	})
})
