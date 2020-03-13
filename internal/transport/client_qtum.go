package transport

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"github.com/hugorut/coins-oracle/pkg/transport"
)

var (
	QtumAssetID = "QTUM"

	qtumDiv    float64 = 100000000
	qtumAPIURL         = "https://qtum.info"
)

// QtumAddressResponse represents a get address JSON response.
type QtumAddressResponse struct {
	Balance          string        `json:"balance"`
	TotalReceived    string        `json:"totalReceived"`
	TotalSent        string        `json:"totalSent"`
	Unconfirmed      string        `json:"unconfirmed"`
	Staking          string        `json:"staking"`
	Mature           string        `json:"mature"`
	Qrc20Balances    []interface{} `json:"qrc20Balances"`
	Qrc721Balances   []interface{} `json:"qrc721Balances"`
	Ranking          int           `json:"ranking"`
	TransactionCount int           `json:"transactionCount"`
	BlocksMined      int           `json:"blocksMined"`
}

// QtumTransactionResponse represents a get tx JSON response.
type QtumTransactionResponse struct {
	ID        string `json:"id"`
	Hash      string `json:"hash"`
	Version   int    `json:"version"`
	LockTime  int    `json:"lockTime"`
	BlockHash string `json:"blockHash"`
	Inputs    []struct {
		PrevTxID    string `json:"prevTxId"`
		OutputIndex int    `json:"outputIndex"`
		Value       string `json:"value"`
		Address     string `json:"address"`
		ScriptSig   struct {
			Type string `json:"type"`
			Hex  string `json:"hex"`
			Asm  string `json:"asm"`
		} `json:"scriptSig"`
		Sequence int64 `json:"sequence"`
	} `json:"inputs"`
	Outputs []struct {
		Value        string `json:"value"`
		Address      string `json:"address"`
		ScriptPubKey struct {
			Type string `json:"type"`
			Hex  string `json:"hex"`
			Asm  string `json:"asm"`
		} `json:"scriptPubKey"`
		SpentTxID  string `json:"spentTxId,omitempty"`
		SpentIndex int    `json:"spentIndex,omitempty"`
	} `json:"outputs"`
	IsCoinbase    bool   `json:"isCoinbase"`
	IsCoinstake   bool   `json:"isCoinstake"`
	BlockHeight   int    `json:"blockHeight"`
	Confirmations int64  `json:"confirmations"`
	Timestamp     int    `json:"timestamp"`
	InputValue    string `json:"inputValue"`
	OutputValue   string `json:"outputValue"`
	RefundValue   string `json:"refundValue"`
	Fees          string `json:"fees"`
	Size          int    `json:"size"`
	Weight        int    `json:"weight"`
}

// QtumBlocksResponse represents a get blocks JSON response.
type QtumBlocksResponse []struct {
	Hash             string `json:"hash"`
	Height           int    `json:"height"`
	Timestamp        int    `json:"timestamp"`
	Interval         int    `json:"interval"`
	Size             int    `json:"size"`
	TransactionCount int    `json:"transactionCount"`
	Miner            string `json:"miner"`
	Reward           string `json:"reward"`
}

// QtumClient is the Qtum implementation of the CoinClient
type QtumClient struct {
	transport.BaseClient
}

// NewQtumClient returns a new client using os variables.
func NewQtumClient() (*QtumClient, error) {
	u, err := url.Parse(qtumAPIURL)
	if err != nil {
		return nil, err
	}

	return &QtumClient{
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
func (b QtumClient) GetInfo() (*transport.CoinState, error) {
	var blocks QtumBlocksResponse
	err := b.GET("/api/blocks", map[string]string{
		"date": time.Now().Format("2006-01-02"),
	}, &blocks)
	if err != nil {
		return nil, errors.Wrap(err, "error making blocks request")
	}

	return &transport.CoinState{
		Data: transport.CoinData{
			Chain:        "main",
			BlockHeight:  blocks[0].Height,
			CurrentBlock: blocks[0].Hash,
		},
	}, nil
}

// GetBalance returns the balance of the address.
func (b QtumClient) GetBalance(addr string) (*transport.Balance, error) {
	var wallet QtumAddressResponse
	err := b.GET("/api/address/"+addr, nil, &wallet)
	if err != nil {
		return nil, errors.Wrap(err, "error making address request")
	}

	return &transport.Balance{
		Data: transport.BalanceData{
			Assets: []transport.Asset{
				{
					Asset:   QtumAssetID,
					Balance: wallet.Balance,
				},
			},
		},
	}, nil
}

// GetTransactionByHash returns the transaction stored at the given hash.
func (b QtumClient) GetTransactionByHash(hash string) (*transport.TransactionResp, error) {
	var transaction QtumTransactionResponse
	err := b.GET("/api/tx/"+hash, nil, &transaction)
	if err != nil {
		return nil, errors.Wrap(err, "error making transaction request")
	}

	output, _ := strconv.ParseFloat(transaction.OutputValue, 64)
	output = output / qtumDiv
	return &transport.TransactionResp{
		Data: struct {
			Transaction transport.Transaction `json:"transaction"`
		}{
			Transaction: transport.Transaction{
				ID:    hash,
				From:  transaction.Inputs[0].Address,
				To:    transaction.Outputs[0].Address,
				Value: fmt.Sprintf("%f", output),
				Confirmations: transport.Confirmations{
					Threshold: transport.ConfirmThresholdValue,
					Confirmed: *transport.ConfirmThresholdValue < transaction.Confirmations,
					Value:     transport.NewInt64(transaction.Confirmations),
				},
			},
		},
	}, nil
}
