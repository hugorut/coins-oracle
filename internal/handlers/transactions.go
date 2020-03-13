package handlers

import (
	"fmt"
	"github.com/hugorut/coins-oracle/pkg/transport"
	"net/http"

	"github.com/labstack/echo"
)

// GetTransactionByHash fetches information about a transaction on a ledger by its hash.
func GetTransactionByHash(c echo.Context) error {
	c.Logger().Print("executing GetTransactionByHash handler")

	hash := c.Param("txHash")
	client := c.Get("coin_client").(transport.CoinClient)

	tr, err := client.GetTransactionByHash(hash)
	if err != nil {
		c.Logger().Errorf("error getting transaction for hash: %s for coin: %s, err: %v", hash, c.Param("assetId"), err)
		return c.JSON(http.StatusBadRequest, genericResponse{
			Error: fmt.Sprintf("could not return transaction details for the given hash/id"),
			Code:  ErrorCodeGetTransactionError,
		})
	}

	return c.JSON(http.StatusOK, tr)
}
