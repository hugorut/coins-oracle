package transport

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/hugorut/coins-oracle/pkg/transport"
)

var (
	DecredAssetID = "DCR"

	decredExplorerAPIURL, _ = url.Parse("https://explorer.dcrdata.org")
)

// DecredTXResponse represents a successful JSON transaction response.
type DecredTXResponse struct {
	Txid     string `json:"txid"`
	Version  int    `json:"version"`
	Locktime int    `json:"locktime"`
	Vin      []struct {
		Txid      string `json:"txid"`
		Vout      int    `json:"vout"`
		Sequence  int64  `json:"sequence"`
		N         int    `json:"n"`
		ScriptSig struct {
			Hex string `json:"hex"`
			Asm string `json:"asm"`
		} `json:"scriptSig"`
		Addr     string  `json:"addr"`
		ValueSat int     `json:"valueSat"`
		Value    float64 `json:"value"`
	} `json:"vin"`
	Vout []struct {
		Value        float64 `json:"value"`
		N            int     `json:"n"`
		ScriptPubKey struct {
			Hex       string   `json:"hex"`
			Asm       string   `json:"asm"`
			Addresses []string `json:"addresses"`
			Type      string   `json:"type"`
		} `json:"scriptPubKey"`
		SpentTxID   string `json:"spentTxId"`
		SpentIndex  int    `json:"spentIndex"`
		SpentHeight int    `json:"spentHeight"`
	} `json:"vout"`
	Blockhash     string  `json:"blockhash"`
	Blockheight   int     `json:"blockheight"`
	Confirmations int64   `json:"confirmations"`
	Time          int     `json:"time"`
	Blocktime     int     `json:"blocktime"`
	ValueOut      float64 `json:"valueOut"`
	Size          int     `json:"size"`
	ValueIn       float64 `json:"valueIn"`
	Fees          float64 `json:"fees"`
}

// DecredBlocksResponse represents a successful JSON blocks response.
type DecredBlocksResponse struct {
	Blocks []struct {
		Height     int    `json:"height"`
		Size       int    `json:"size"`
		Hash       string `json:"hash"`
		Unixtime   string `json:"unixtime"`
		Txlength   int    `json:"txlength"`
		Voters     int    `json:"voters"`
		Freshstake int    `json:"freshstake"`
	} `json:"blocks"`
	Length     int `json:"length"`
	Pagination struct {
		Next      string `json:"next"`
		Prev      string `json:"prev"`
		CurrentTs int    `json:"currentTs"`
		Current   string `json:"current"`
		IsToday   bool   `json:"isToday"`
		More      bool   `json:"more"`
		MoreTs    int    `json:"moreTs"`
	} `json:"pagination"`
}

// DecredAddressResponse represents a successful address response.
type DecredAddressResponse struct {
	AddrStr                 string  `json:"addrStr"`
	Balance                 float64 `json:"balance"`
	BalanceSat              int64   `json:"balanceSat"`
	TotalReceived           float64 `json:"totalReceived"`
	TotalReceivedSat        int64   `json:"totalReceivedSat"`
	TotalSent               float64 `json:"totalSent"`
	TotalSentSat            int64   `json:"totalSentSat"`
	UnconfirmedBalance      int     `json:"unconfirmedBalance"`
	UnconfirmedBalanceSat   int     `json:"unconfirmedBalanceSat"`
	UnconfirmedTxApperances int     `json:"unconfirmedTxApperances"`
	TxApperances            int     `json:"txApperances"`
}

// DecredClient is the Decred implementation of the CoinClient
type DecredClient struct {
	transport.BaseClient
}

// NewDecredClient returns a new client using os variables.
func NewDecredClient() (*DecredClient, error) {
	return &DecredClient{
		BaseClient: transport.BaseClient{
			BaseURL: decredExplorerAPIURL,
			Client: &http.Client{
				Timeout: transport.DefaultClientTimeout,
			},
			Log: transport.StdLogger,
		},
	}, nil
}

// GetInfo attempts to get standardised coin info from multiple rpc calls.
func (d DecredClient) GetInfo() (*transport.CoinState, error) {
	var info DecredBlocksResponse

	if err := d.GET("/insight/api/blocks", map[string]string{"limit": "1"}, &info); err != nil {
		return nil, err
	}

	mostRecentBlock := info.Blocks[0]

	return &transport.CoinState{
		Data: transport.CoinData{
			Chain:        "main",
			BlockHeight:  mostRecentBlock.Height,
			CurrentBlock: mostRecentBlock.Hash,
		},
	}, nil
}

// GetBalance returns the balance of the address.
func (d DecredClient) GetBalance(addr string) (*transport.Balance, error) {
	var account DecredAddressResponse

	if err := d.GET("/insight/api/addr/"+addr, map[string]string{"noTxList": "1"}, &account); err != nil {
		return nil, err
	}

	return &transport.Balance{
		Data: transport.BalanceData{
			Assets: []transport.Asset{
				{
					Asset:   DecredAssetID,
					Balance: fmt.Sprintf("%f", account.Balance),
				},
			},
		},
	}, nil
}

// GetTransactionByHash returns the transaction stored at the given hash.
func (d DecredClient) GetTransactionByHash(hash string) (*transport.TransactionResp, error) {
	var tx DecredTXResponse

	if err := d.GET("/insight/api/tx/"+hash, nil, &tx); err != nil {
		return nil, err
	}

	return &transport.TransactionResp{
		Data: struct {
			Transaction transport.Transaction `json:"transaction"`
		}{
			Transaction: transport.Transaction{
				ID:    hash,
				From:  tx.Vin[0].Addr,
				To:    tx.Vout[0].ScriptPubKey.Addresses[0],
				Value: fmt.Sprintf("%f", tx.ValueOut),
				Confirmations: transport.Confirmations{
					Threshold: transport.ConfirmThresholdValue,
					Confirmed: tx.Confirmations > *transport.ConfirmThresholdValue,
					Value:     transport.NewInt64(tx.Confirmations),
				},
			},
		},
	}, nil
}
