package transport

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/stellar/go/support/errors"

	"github.com/hugorut/coins-oracle/pkg/transport"
)

var (
	WavesAssetID = "WAVES"
)

// WavesGetBalanceResponse represents the json returned from a balance/details request.
type WavesGetBalanceResponse struct {
	Address    string `json:"address"`
	Regular    int64  `json:"regular"`
	Generating int64  `json:"generating"`
	Available  int64  `json:"available"`
	Effective  int64  `json:"effective"`
}

// WavesGetBlockResponse represents the json returned from a blocks latest call.
type WavesGetBlockResponse struct {
	Blocksize    int           `json:"blocksize"`
	Reward       int           `json:"reward"`
	Signature    string        `json:"signature"`
	Fee          int           `json:"fee"`
	Generator    string        `json:"generator"`
	Transactions []interface{} `json:"transactions"`
	Version      int           `json:"version"`
	Reference    string        `json:"reference"`
	Features     []interface{} `json:"features"`
	TotalFee     int           `json:"totalFee"`
	NxtConsensus struct {
		BaseTarget          int    `json:"base-target"`
		GenerationSignature string `json:"generation-signature"`
	} `json:"nxt-consensus"`
	DesiredReward    int   `json:"desiredReward"`
	TransactionCount int   `json:"transactionCount"`
	Timestamp        int64 `json:"timestamp"`
	Height           int   `json:"height"`
}

// WavesGetTXResponse represents the json returned from a successful transactions/info call.
type WavesGetTXResponse struct {
	SenderPublicKey string      `json:"senderPublicKey"`
	Amount          int64       `json:"amount"`
	Signature       string      `json:"signature"`
	Fee             int         `json:"fee"`
	Type            int         `json:"type"`
	Version         int         `json:"version"`
	Attachment      string      `json:"attachment"`
	Sender          string      `json:"sender"`
	FeeAssetID      interface{} `json:"feeAssetId"`
	Proofs          []string    `json:"proofs"`
	AssetID         interface{} `json:"assetId"`
	Recipient       string      `json:"recipient"`
	FeeAsset        interface{} `json:"feeAsset"`
	ID              string      `json:"id"`
	Timestamp       int64       `json:"timestamp"`
	Height          int         `json:"height"`
}

// WavesClient is the Waves implementation of the CoinClient
type WavesClient struct {
	transport.BaseClient
}

// NewWavesClient returns a new client using os variables.
func NewWavesClient() (*WavesClient, error) {
	u, err := url.Parse(getNodeURL("WAVES_URL"))
	if err != nil {
		return nil, err
	}

	return &WavesClient{
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
func (w WavesClient) GetInfo() (*transport.CoinState, error) {
	var res WavesGetBlockResponse

	err := w.GET("/blocks/last", nil, &res)
	if err != nil {
		return nil, errors.Wrap(err, "error getting latest waves block")
	}

	return &transport.CoinState{
		Data: transport.CoinData{
			Chain:        "main",
			BlockHeight:  res.Height,
			CurrentBlock: fmt.Sprintf("%d", res.Height),
		},
	}, nil
}

// GetBalance returns the balance of the address.
func (w WavesClient) GetBalance(addr string) (*transport.Balance, error) {
	var res WavesGetBalanceResponse

	err := w.GET("/addresses/balance/details/"+addr, nil, &res)
	if err != nil {
		return nil, errors.Wrap(err, "error getting waves balance for address")
	}

	return &transport.Balance{
		Data: transport.BalanceData{
			Assets: []transport.Asset{
				{
					Asset:   WavesAssetID,
					Balance: fmt.Sprintf("%d", res.Regular),
				},
			},
		},
	}, nil
}

// GetTransactionByHash returns the transaction stored at the given hash.
func (w WavesClient) GetTransactionByHash(hash string) (*transport.TransactionResp, error) {
	var res WavesGetTXResponse

	err := w.GET("/transactions/info/"+hash, nil, &res)
	if err != nil {
		return nil, errors.Wrap(err, "error getting waves transaction from hash")
	}

	return &transport.TransactionResp{
		Data: struct {
			Transaction transport.Transaction `json:"transaction"`
		}{
			Transaction: transport.Transaction{
				ID:    hash,
				From:  res.Sender,
				To:    res.Recipient,
				Value: fmt.Sprintf("%d", res.Amount),
				Confirmations: transport.Confirmations{
					// if the transaction is returned from this endpoint then it is confirmed
					Confirmed: true,
				},
			},
		},
	}, nil
}
