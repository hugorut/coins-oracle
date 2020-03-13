package transport

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/hugorut/coins-oracle/pkg/transport"
)

var (
	NanoAssetID = "NANO"
)

// NanoBlockCountResponse is a struct representing the json from a successful block_count call.
type NanoBlockCountResponse struct {
	Count     string `json:"count"`
	Unchecked string `json:"unchecked"`
}

// NanoBlockResponse is a struct representing the json from a successful block_info call.
type NanoAccountResponse struct {
	Frontier            string `json:"frontier"`
	OpenBlock           string `json:"open_block"`
	RepresentativeBlock string `json:"representative_block"`
	Balance             string `json:"balance"`
	ModifiedTimestamp   string `json:"modified_timestamp"`
	BlockCount          string `json:"block_count"`
	AccountVersion      string `json:"account_version"`
	ConfirmationHeight  string `json:"confirmation_height"`
}

// NanoBlockResponse is a struct representing the json from a successful block_info call.
type NanoBlockResponse struct {
	Account        string    `json:"account"`
	Amount         string    `json:"amount"`
	Type           string    `json:"type"`
	Representative string    `json:"representative"`
	Previous       string    `json:"previous"`
	Work           string    `json:"work"`
	Signature      string    `json:"signature"`
	Date           time.Time `json:"date"`
	Link           string    `json:"link"`
	LinkAsAccount  string    `json:"link_as_account"`
	Balance        string    `json:"balance"`
}

// NanoBlockInfoRequest is a struct to hold the block_info json action request.
type NanoBlockInfoRequest struct {
	Action    string `json:"action"`
	JSONBlock string `json:"json_block"`
	Hash      string `json:"hash"`
}

// NanoAccountInfoRequest is a struct to hold the account_info json action request.
type NanoAccountInfoRequest struct {
	Action  string `json:"action"`
	Account string `json:"account"`
}

// NanoActionRequest is a struct to hold the base json action request.
type NanoActionRequest struct {
	Action string `json:"action"`
}

// NanoClient is the Nano implementation of the CoinClient
type NanoClient struct {
	transport.BaseClient
}

// NewNanoClient returns a new client using os variables.
func NewNanoClient() (*NanoClient, error) {
	u, err := url.Parse(getNodeURL("NANO_URL"))
	if err != nil {
		return nil, err
	}

	return &NanoClient{
		BaseClient: transport.BaseClient{
			BaseURL: u,
			Client: &http.Client{
				Timeout: transport.DefaultClientTimeout,
			},
			Log: transport.StdLogger,
		},
	}, nil
}

// GetInfo attempts to get standardised coin info from multiple rpc calls.
func (n NanoClient) GetInfo() (*transport.CoinState, error) {
	var info NanoBlockCountResponse

	if err := n.POST(NanoActionRequest{Action: "block_count"}, "/", &info); err != nil {
		return nil, err
	}

	height, _ := strconv.Atoi(info.Count)
	return &transport.CoinState{
		Data: transport.CoinData{
			Chain: "main",
			// Nano's API has no way of getting the current block hash, so the count will have to do.
			CurrentBlock: info.Count,
			BlockHeight:  height,
		},
	}, nil
}

// GetBalance returns the balance of the address.
func (n NanoClient) GetBalance(addr string) (*transport.Balance, error) {
	var acc NanoAccountResponse

	if err := n.POST(NanoAccountInfoRequest{Action: "account_info", Account: addr}, "/", &acc); err != nil {
		return nil, err
	}

	return &transport.Balance{
		Data: transport.BalanceData{
			Assets: []transport.Asset{
				{
					Asset:   NanoAssetID,
					Balance: acc.Balance,
				},
			},
		},
	}, nil
}

// GetTransactionByHash returns the transaction stored at the given hash.
func (n NanoClient) GetTransactionByHash(hash string) (*transport.TransactionResp, error) {
	var block NanoBlockResponse

	if err := n.POST(NanoBlockInfoRequest{
		Action:    "block_info",
		JSONBlock: "true",
		Hash:      hash,
	}, "/", &block); err != nil {
		return nil, err
	}

	return &transport.TransactionResp{
		Data: struct {
			Transaction transport.Transaction `json:"transaction"`
		}{
			Transaction: transport.Transaction{
				ID:    hash,
				From:  block.Account,
				To:    block.LinkAsAccount,
				Value: block.Amount,
				Confirmations: transport.Confirmations{
					// if block is present in the local node it is confirmed
					Confirmed: true,
				},
			},
		},
	}, nil
}
