package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/hugorut/coins-oracle/pkg/transport"
	"net/http"

	"github.com/labstack/echo"
)

// GetWalletBalance fetches the current balance of assets in the address.
func GetWalletBalance(c echo.Context) error {
	c.Logger().Print("executing GetWalletBalance handler")

	addr := c.Param("addr")
	client := c.Get("coin_client").(transport.CoinClient)

	ob, err := client.GetBalance(addr)
	if err != nil {
		c.Logger().Errorf("error getting balance for wallet address: %s for coin: %s, err: %v", addr, c.Param("assetId"), err)
		return c.JSON(http.StatusBadRequest, genericResponse{
			Error: fmt.Sprintf("could not get balance of given address"),
			Code:  ErrorCodeBalanceError,
		})
	}

	return c.JSON(http.StatusOK, ob)
}

type importAddressReq struct {
	Addr string `json:"addr"`
}

// ImportAddress tells the asset to index a specific address so that wallet functionality can occur in the future.
func ImportAddress(c echo.Context) error {
	c.Logger().Print("executing ImportAddress handler")

	var req importAddressReq
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return c.JSON(http.StatusBadRequest, genericResponse{
			Error: "missing addr field in request",
			Code:  ErrorInvalidRequest,
		})
	}

	client, ok := c.Get("coin_client").(transport.AddressImporter)
	if !ok {
		return c.JSON(http.StatusBadRequest, genericResponse{
			Error: fmt.Sprintf("client: %s does not have import address functionality", c.Param("asset_id")),
			Code:  ErrorCodeCannotImport,
		})
	}

	err := client.ImportAddress(req.Addr)
	if err != nil {
		c.Logger().Errorf("error getting importing address: %s for coin: %s, err: %v", req.Addr, c.Param("assetId"), err)
		return c.JSON(http.StatusBadRequest, genericResponse{
			Error: "could not import address",
			Code:  ErrorCodeCannotImport,
		})
	}

	return c.JSON(http.StatusOK, successResponse)
}
