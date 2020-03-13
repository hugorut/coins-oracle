package handlers

import (
	"fmt"
	transport2 "github.com/hugorut/coins-oracle/pkg/transport"
	"net/http"
	"strings"

	"github.com/labstack/echo"

	"github.com/hugorut/coins-oracle/internal/transport"
)

// GetNodesResponse struct to map the coins info to the required json format.
type GetNodesResponse struct {
	Data struct {
		Nodes []transport.CoinNode `json:"nodes"`
	} `json:"data"`
}

// GetInfo fetches information about a specific node.
func GetInfo(c echo.Context) error {
	c.Logger().Print("executing GetInfo handler")
	client := c.Get("coin_client").(transport2.CoinClient)

	info, err := client.GetInfo()
	if err != nil {
		c.Logger().Errorf("error getting info for coin: %s, err: %v", c.Param("assetId"), err)
		return c.JSON(http.StatusBadRequest, genericResponse{
			Error: fmt.Sprintf("unable to get node information for given coin"),
			Code:  ErrorCodeGetInfoError,
		})
	}

	return c.JSON(http.StatusOK, info)
}

// GetNodes fetches a list of all available nodes.
func GetNodes(c echo.Context) error {
	c.Logger().Printf("executing GetNodes handler")
	router := c.Get("coin_router").(transport.Resolver)

	noinfo := strings.ToLower(c.QueryParam("noinfo")) == "true"
	coins := router.GetNodes(!noinfo)

	return c.JSON(http.StatusOK, GetNodesResponse{
		Data: struct {
			Nodes []transport.CoinNode `json:"nodes"`
		}{
			Nodes: coins,
		},
	})
}
