package handlers_test

import (
	"fmt"
	transport2 "github.com/hugorut/coins-oracle/pkg/transport"
	"net/http"
	"net/http/httptest"
	"sync"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/hugorut/coins-oracle/internal/handlers"
	"github.com/hugorut/coins-oracle/internal/transport"
	mock_transport "github.com/hugorut/coins-oracle/internal/transport/mocks"
)

var _ = Describe("Handler", func() {
	var (
		e *echo.Echo
	)

	BeforeEach(func() {
		e = echo.New()
	})

	Describe("H", func() {
		Describe("Ping", func() {
			It("Should return pong message when executed", func() {
				req := httptest.NewRequest(http.MethodGet, "/ping", nil)
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
				rec := httptest.NewRecorder()

				c := e.NewContext(req, rec)

				Expect(Ping(c)).To(Succeed())
				Expect(rec.Body.String()).Should(MatchJSON(`{
					"message": "pong"
				}`))
			})
		})
	})

	Describe("SetCoinClientMiddlewareFunc", func() {
		var (
			r      *transport.CoinResolver
			ctrl   *gomock.Controller
			client transport2.CoinClient
		)

		BeforeEach(func() {
			r = &transport.CoinResolver{
				C:  make(map[string]transport2.CoinClient),
				Mu: &sync.Mutex{},
			}

			ctrl = gomock.NewController(GinkgoT())
			client = mock_transport.NewMockCoinClient(ctrl)
		})

		AfterEach(func() {
			ctrl.Finish()
		})

		Context("With valid assetId", func() {
			It("Should set a CoinClient in the request", func() {
				assetID := "test-coin"
				r.Register(assetID, client)

				req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/nodes/%s", assetID), nil)
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
				rec := httptest.NewRecorder()

				c := e.NewContext(req, rec)
				c.SetParamNames("assetId")
				c.SetParamValues(assetID)

				f := SetCoinClientMiddlewareFunc(r)
				err := f(func(c echo.Context) error {
					v := c.Get("coin_client")

					Expect(v).To(BeIdenticalTo(client))
					return nil
				})(c)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("With invalid assetId", func() {
			It("Should terminate middleware chain with error", func() {
				assetID := "test-coin"

				req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/nodes/%s", assetID), nil)
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
				rec := httptest.NewRecorder()

				c := e.NewContext(req, rec)
				c.SetParamNames("assetId")
				c.SetParamValues(assetID)

				f := SetCoinClientMiddlewareFunc(r)
				err := f(func(c echo.Context) error {
					return nil
				})(c)
				Expect(err).ToNot(HaveOccurred())

				Expect(rec.Body.String()).Should(MatchJSON(`{
					"error": "asset: test-coin was not found"
				}`))
				Expect(rec.Code).To(Equal(http.StatusNotFound))
			})
		})

		Context("On non nodes route", func() {
			It("Should not set client", func() {
				assetID := "somehash"

				req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/some-other-route/%s", assetID), nil)
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
				rec := httptest.NewRecorder()

				c := e.NewContext(req, rec)
				c.SetParamNames("assetId")
				c.SetParamValues(assetID)

				f := SetCoinClientMiddlewareFunc(r)
				err := f(func(c echo.Context) error {
					v := c.Get("coin_client")

					Expect(v).To(BeNil())
					return nil
				})(c)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe("SetRouterMiddlewareFunc", func() {
		var (
			r *transport.CoinResolver
		)

		BeforeEach(func() {
			r = &transport.CoinResolver{
				C:  make(map[string]transport2.CoinClient),
				Mu: &sync.Mutex{},
			}
		})

		It("Should set router in context", func() {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			f := SetRouterMiddlewareFunc(r)

			err := f(func(c echo.Context) error {
				v := c.Get("coin_router")

				Expect(v).To(BeIdenticalTo(r))
				return nil
			})(c)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
