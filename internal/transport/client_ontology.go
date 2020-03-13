package transport

import (
	"fmt"
	"net/http"
	"net/url"

	ontology_go_sdk "github.com/ontio/ontology-go-sdk"
	"github.com/ontio/ontology/common"
	"github.com/pkg/errors"

	"github.com/hugorut/coins-oracle/pkg/transport"
)

var (
	OntologyAssetID = "ONT"
)

// ONTGetResultResponse represents a generic success response from ont.
type ONTGetResultResponse struct {
	Action  string      `json:"Action"`
	Desc    string      `json:"Desc"`
	Error   int         `json:"Error"`
	Result  interface{} `json:"Result"`
	Version string      `json:"Version"`
}

// ONTGetTransactionResponse represents a transaction response.
type ONTGetTransactionResponse struct {
	Action string `json:"Action"`
	Desc   string `json:"Desc"`
	Error  int    `json:"Error"`
	Result struct {
		Version  int    `json:"Version"`
		Nonce    int    `json:"Nonce"`
		GasPrice int    `json:"GasPrice"`
		GasLimit int    `json:"GasLimit"`
		Payer    string `json:"Payer"`
		TxType   int    `json:"TxType"`
		Payload  struct {
			Code string `json:"Code"`
		} `json:"Payload"`
		Attributes []interface{} `json:"Attributes"`
		Sigs       []struct {
			PubKeys []string `json:"PubKeys"`
			M       int      `json:"M"`
			SigData []string `json:"SigData"`
		} `json:"Sigs"`
		Hash   string `json:"Hash"`
		Height int    `json:"Height"`
	} `json:"Result"`
	Version string `json:"Version"`
}

// ONTGetBalanceResponse represents a json balance response
type ONTGetBalanceResponse struct {
	Action string `json:"Action"`
	Desc   string `json:"Desc"`
	Error  int    `json:"Error"`
	Result struct {
		Ont string `json:"ont"`
		Ong string `json:"ong"`
	} `json:"Result"`
	Version string `json:"Version"`
}

// OntologyClient is the Ontology implementation of the CoinClient
type OntologyClient struct {
	transport.BaseClient
}

// NewOntologyClient returns a new client using os variables.
func NewOntologyClient() (*OntologyClient, error) {
	u, err := url.Parse(getNodeURL("ONTOLOGY_URL"))
	if err != nil {
		return nil, err
	}

	return &OntologyClient{
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
func (b OntologyClient) GetInfo() (*transport.CoinState, error) {
	var height ONTGetResultResponse
	if err := b.GET("/api/v1/block/height", nil, &height); err != nil {
		return nil, errors.Wrap(err, "error getting ontology block height")
	}

	var hash ONTGetResultResponse
	h := int(height.Result.(float64))
	if err := b.GET(fmt.Sprintf("/api/v1/block/hash/%d", h), nil, &hash); err != nil {
		return nil, errors.Wrap(err, "error getting ontology block hash")
	}

	return &transport.CoinState{
		Data: transport.CoinData{
			Chain:        "main",
			BlockHeight:  h,
			CurrentBlock: fmt.Sprintf("%s", hash.Result),
		},
	}, nil
}

// GetBalance returns the balance of the address.
func (b OntologyClient) GetBalance(addr string) (*transport.Balance, error) {
	var balance ONTGetBalanceResponse
	if err := b.GET("/api/v1/balance/"+addr, nil, &balance); err != nil {
		return nil, errors.Wrap(err, "error getting ontology balance for account")
	}

	return &transport.Balance{
		Data: transport.BalanceData{
			Assets: []transport.Asset{
				{
					Asset:   OntologyAssetID,
					Balance: balance.Result.Ont,
				},
			},
		},
	}, nil
}

// GetTransactionByHash returns the transaction stored at the given hash.
func (b OntologyClient) GetTransactionByHash(hash string) (*transport.TransactionResp, error) {
	var tx ONTGetTransactionResponse
	if err := b.GET("/api/v1/transaction/"+hash, nil, &tx); err != nil {
		return nil, errors.Wrap(err, "error getting ontology transaction by hash")
	}

	payloadHex := tx.Result.Payload.Code
	byt, err := common.HexToBytes(payloadHex)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding ontology payload for transaction")
	}

	payload, err := ontology_go_sdk.ParsePayload(byt)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding ontology payload for transaction")
	}

	return &transport.TransactionResp{
		Data: struct {
			Transaction transport.Transaction `json:"transaction"`
		}{
			Transaction: transport.Transaction{
				ID:    tx.Result.Hash,
				From:  payload["from"].(string),
				To:    payload["to"].(string),
				Value: fmt.Sprintf("%v", payload["amount"]),
				Confirmations: transport.Confirmations{
					Confirmed: true,
				},
			},
		},
	}, nil
}
