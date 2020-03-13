package handlers

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/labstack/echo"

	"github.com/hugorut/coins-oracle/internal/transport"
)

const (
	ErrorInvalidRequest = 101

	ErrorCodeCannotImport = 201
	ErrorCodeBalanceError = 202

	ErrorCodeGetTransactionError = 301

	ErrorCodeGetInfoError = 401
)

var (
	coinsReg = regexp.MustCompile(`^/nodes/`)

	successResponse = genericResponse{Data: "success"}
)

type genericResponse struct {
	Data  interface{} `json:"data"`
	Error string      `json:"error"`
	Code  int         `json:"code"`
}

// SetRouterMiddlewareFunc applies a router to the context.
func SetRouterMiddlewareFunc(router transport.Resolver) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("coin_router", router)

			return next(c)
		}
	}
}

// SetCoinClientMiddlewareFunc returns a middleware func using the router provided
// to rectify the asset in the request
func SetCoinClientMiddlewareFunc(router transport.Resolver) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			assetID := c.Param("assetId")

			if !coinsReg.MatchString(c.Request().URL.Path) || assetID == "" {
				return next(c)
			}

			client, err := router.Get(assetID)
			if err != nil {
				return c.JSON(http.StatusNotFound, map[string]string{
					"error": fmt.Sprintf("asset: %s was not found", assetID),
				})
			}

			c.Set("coin_client", client)
			return next(c)
		}
	}
}

// Ping provides a utility function to make sure the lambda is up.
// Ping the handler every 5s reduces the cold startup time.
func Ping(c echo.Context) error {
	return c.JSON(200, map[string]string{
		"message": "pong",
	})
}
