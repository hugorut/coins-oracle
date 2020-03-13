package transport

import (
	"net/http"
	"net/url"

	"github.com/hugorut/coins-oracle/pkg/transport"
)

var (
	NeoAssetID = "NEO"
)

// NeoRPCRequest represents the JSON needed to make a request to the neo RPC API.
type NeoRPCRequest struct {
	JsonRPC string   `json:"jsonrpc"`
	Method  string   `json:"method"`
	Params  []string `json:"params"`
	ID      int      `json:"id"`
}

// NeoAccountResponse represents the JSON returned from a successful get account call.
type NeoAccountResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  struct {
		Version    int           `json:"version"`
		ScriptHash string        `json:"script_hash"`
		Frozen     bool          `json:"frozen"`
		Votes      []interface{} `json:"votes"`
		Balances   []struct {
			Asset string `json:"asset"`
			Value string `json:"value"`
		} `json:"balances"`
	} `json:"result"`
}

// NeoBlockHashResponse represents the JSON returned from a successful get latest block hash call.
type NeoBlockHashResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  string `json:"result"`
}

// NeoBlockCountResponse represents the JSON returned form a successful block count call.
type NeoBlockCountResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  int    `json:"result"`
}

// NeoTXResponse represents the JSON returned from a successful get transaction call.
type NeoTXResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  struct {
		Txid       string        `json:"Txid"`
		Size       int           `json:"Size"`
		Type       string        `json:"Type"`
		Version    int           `json:"Version"`
		Attributes []interface{} `json:"Attributes"`
		Vin        []struct {
			Txid string `json:"Txid"`
			Vout int    `json:"Vout"`
		} `json:"Vin"`
		Vout []struct {
			N       int    `json:"N"`
			Asset   string `json:"Asset"`
			Value   string `json:"Value"`
			Address string `json:"Address"`
		} `json:"Vout"`
		SysFee  string `json:"Sys_fee"`
		NetFee  string `json:"Net_fee"`
		Scripts []struct {
			Invocation   string `json:"Invocation"`
			Verification string `json:"Verification"`
		} `json:"Scripts"`
		Blockhash     string `json:"Blockhash"`
		Confirmations int64  `json:"Confirmations"`
		Blocktime     int    `json:"Blocktime"`
	} `json:"result"`
}

// NeoClient is the Neo implementation of the CoinClient
type NeoClient struct {
	transport.BaseClient
}

// NewNeoClient returns a new client using os variables.
func NewNeoClient() (*NeoClient, error) {
	u, err := url.Parse(getNodeURL("NEO_URL"))
	if err != nil {
		return nil, err
	}

	return &NeoClient{
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
func (n NeoClient) GetInfo() (*transport.CoinState, error) {
	var info NeoBlockHashResponse

	req := NeoRPCRequest{
		JsonRPC: transport.RPCVersion,
		Params:  []string{},
		Method:  "getbestblockhash",
		ID:      1,
	}

	if err := n.POST(req, "/", &info); err != nil {
		return nil, err
	}

	var count NeoBlockCountResponse
	req.Method = "getblockcount"
	req.ID = 2

	if err := n.POST(req, "/", &count); err != nil {
		return nil, err
	}

	return &transport.CoinState{
		Data: transport.CoinData{
			Chain:        "main",
			BlockHeight:  count.Result,
			CurrentBlock: info.Result,
		},
	}, nil
}

// GetBalance returns the balance of the address.
func (n NeoClient) GetBalance(addr string) (*transport.Balance, error) {
	var info NeoAccountResponse

	req := NeoRPCRequest{
		JsonRPC: transport.RPCVersion,
		Method:  "getaccountstate",
		Params: []string{
			addr,
		},
		ID: 1,
	}

	if err := n.POST(req, "/", &info); err != nil {
		return nil, err
	}

	assets := make([]transport.Asset, len(info.Result.Balances))
	for key, value := range info.Result.Balances {
		assets[key] = transport.Asset{
			Asset:   transport.StripHex(value.Asset),
			Balance: value.Value,
		}
	}

	return &transport.Balance{
		Data: transport.BalanceData{
			Assets: assets,
		},
	}, nil
}

// GetTransactionByHash returns the transaction stored at the given hash.
func (n NeoClient) GetTransactionByHash(hash string) (*transport.TransactionResp, error) {
	var tx NeoTXResponse

	req := NeoRPCRequest{
		JsonRPC: transport.RPCVersion,
		Method:  "getrawtransaction",
		Params: []string{
			hash,
			"1",
		},
		ID: 1,
	}

	if err := n.POST(req, "/", &tx); err != nil {
		return nil, err
	}

	var sendingTx NeoTXResponse

	sender := tx.Result.Vin[0]
	req.Params[0] = transport.StripHex(sender.Txid)

	if err := n.POST(req, "/", &sendingTx); err != nil {
		return nil, err
	}

	return &transport.TransactionResp{
		Data: struct {
			Transaction transport.Transaction `json:"transaction"`
		}{
			Transaction: transport.Transaction{
				ID:    hash,
				From:  sendingTx.Result.Vout[sender.Vout].Address,
				To:    tx.Result.Vout[0].Address,
				Value: tx.Result.Vout[0].Value,
				Confirmations: transport.Confirmations{
					Threshold: transport.ConfirmThresholdValue,
					Confirmed: tx.Result.Confirmations >= *transport.ConfirmThresholdValue,
					Value:     &tx.Result.Confirmations,
				},
			},
		},
	}, nil
}
