package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/echo"
	"github.com/labstack/echo"

	"github.com/hugorut/coins-oracle/internal/handlers"
	"github.com/hugorut/coins-oracle/internal/transport"
)

var (
	echoAdapter *echoadapter.EchoLambda
)

func init() {
	r := echo.New()
	resolver := transport.NewResolver(r.Logger)

	r.GET("/ping", handlers.Ping)

	// set all urls under the nodes prefix to use the coin client middleware function
	// which sets a coin client for the given :assetId if one is provided.
	ng := r.Group(
		"/nodes",
		handlers.SetRouterMiddlewareFunc(resolver),
		handlers.SetCoinClientMiddlewareFunc(resolver),
	)

	// node routes
	ng.GET("", handlers.GetNodes)
	ng.GET("/:assetId/info", handlers.GetInfo)

	// address routes
	ng.GET("/:assetId/addrs/:addr/balance", handlers.GetWalletBalance)
	ng.POST("/:assetId/addrs/import", handlers.ImportAddress)

	// transaction routes
	ng.GET("/:assetId/txs/:txHash", handlers.GetTransactionByHash)

	echoAdapter = echoadapter.New(r)
}

// Handler wraps the echo adapter in a common function that the lambda start accepts
func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return echoAdapter.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(Handler)
}
