package transport

import (
	"net/http"
	"net/url"

	"github.com/pkg/errors"

	"github.com/hugorut/coins-oracle/pkg/transport"
)

var (
	LiskAssetID = "LSK"
)

// LiskMeta is a struct representing the json meta data in a response.
type LiskMeta struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

// LiskGetAccountResponse represents a the json returned from a successful accounts call.
type LiskGetAccountResponse struct {
	Meta LiskMeta `json:"meta"`
	Data []struct {
		Address         string      `json:"address"`
		PublicKey       string      `json:"publicKey"`
		Balance         string      `json:"balance"`
		SecondPublicKey string      `json:"secondPublicKey"`
		Asset           interface{} `json:"asset"`
		Delegate        struct {
			Username       string  `json:"username"`
			Vote           string  `json:"vote"`
			Rewards        string  `json:"rewards"`
			ProducedBlocks int     `json:"producedBlocks"`
			MissedBlocks   int     `json:"missedBlocks"`
			Rank           int     `json:"rank"`
			Productivity   float64 `json:"productivity"`
			Approval       float64 `json:"approval"`
		} `json:"delegate"`
	} `json:"data"`
}

// LiskGetBlocksResponse represents a the json returned from a successful blocks call.
type LiskGetBlocksResponse struct {
	Meta LiskMeta `json:"meta"`
	Data []struct {
		ID                   string `json:"id"`
		Version              int    `json:"version"`
		Timestamp            int    `json:"timestamp"`
		Height               int    `json:"height"`
		NumberOfTransactions int    `json:"numberOfTransactions"`
		TotalAmount          string `json:"totalAmount"`
		TotalFee             string `json:"totalFee"`
		Reward               string `json:"reward"`
		PayloadLength        int    `json:"payloadLength"`
		PayloadHash          string `json:"payloadHash"`
		GeneratorPublicKey   string `json:"generatorPublicKey"`
		BlockSignature       string `json:"blockSignature"`
		Confirmations        int    `json:"confirmations"`
		TotalForged          string `json:"totalForged"`
		GeneratorAddress     string `json:"generatorAddress"`
		PreviousBlockID      string `json:"previousBlockId"`
	} `json:"data"`
}

// LiskGetTXResponse represents a the json returned from a successful transactions call.
type LiskGetTXResponse struct {
	Meta LiskMeta `json:"meta"`
	Data []struct {
		ID                 string        `json:"id"`
		Height             int           `json:"height"`
		BlockID            string        `json:"blockId"`
		Type               int           `json:"type"`
		Timestamp          int           `json:"timestamp"`
		SenderPublicKey    string        `json:"senderPublicKey"`
		RecipientPublicKey string        `json:"recipientPublicKey"`
		SenderID           string        `json:"senderId"`
		RecipientID        string        `json:"recipientId"`
		Amount             string        `json:"amount"`
		Fee                string        `json:"fee"`
		Signature          string        `json:"signature"`
		SignSignature      string        `json:"signSignature"`
		Signatures         []interface{} `json:"signatures"`
		Asset              struct {
		} `json:"asset"`
		Confirmations int64 `json:"confirmations"`
	} `json:"data"`
}

// LiskClient is the Lisk implementation of the CoinClient
type LiskClient struct {
	transport.BaseClient
}

// NewLiskClient returns a new client using os variables.
func NewLiskClient() (*LiskClient, error) {
	u, err := url.Parse(getNodeURL("LISK_URL"))
	if err != nil {
		return nil, err
	}

	return &LiskClient{
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
func (b LiskClient) GetInfo() (*transport.CoinState, error) {
	var res LiskGetBlocksResponse

	err := b.GET("/api/blocks", map[string]string{
		"limit": "1",
		"sort":  "height:desc",
	}, &res)
	if err != nil {
		return nil, errors.Wrap(err, "error getting latest lisk block")
	}

	data := res.Data[0]

	return &transport.CoinState{
		Data: transport.CoinData{
			Chain:        "main",
			BlockHeight:  data.Height,
			CurrentBlock: data.ID,
		},
	}, nil
}

// GetBalance returns the balance of the address.
func (b LiskClient) GetBalance(addr string) (*transport.Balance, error) {
	var res LiskGetAccountResponse

	err := b.GET("/api/accounts", map[string]string{
		"address": addr,
		"limit":   "1",
	}, &res)
	if err != nil {
		return nil, errors.Wrap(err, "error getting latest lisk account")
	}

	return &transport.Balance{
		Data: transport.BalanceData{
			Assets: []transport.Asset{
				{
					Asset:   LiskAssetID,
					Balance: res.Data[0].Balance,
				},
			},
		},
	}, nil
}

// GetTransactionByHash returns the transaction stored at the given hash.
func (b LiskClient) GetTransactionByHash(hash string) (*transport.TransactionResp, error) {
	var res LiskGetTXResponse

	err := b.GET("/api/transactions", map[string]string{
		"id":    hash,
		"limit": "1",
	}, &res)
	if err != nil {
		return nil, errors.Wrap(err, "error getting latest lisk transaction")
	}

	tx := res.Data[0]
	return &transport.TransactionResp{
		Data: struct {
			Transaction transport.Transaction `json:"transaction"`
		}{
			Transaction: transport.Transaction{
				ID:    hash,
				From:  tx.SenderID,
				To:    tx.RecipientID,
				Value: tx.Amount,
				Confirmations: transport.Confirmations{
					Threshold: transport.ConfirmThresholdValue,
					Confirmed: tx.Confirmations >= *transport.ConfirmThresholdValue,
					Value:     &tx.Confirmations,
				},
			},
		},
	}, nil
}
