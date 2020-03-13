package transport

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/hugorut/coins-oracle/pkg/transport"
)

var (
	NemAssetID = "XEM"
)

type NemGetLastBlockResponse struct {
	TimeStamp     int    `json:"timeStamp"`
	Signature     string `json:"signature"`
	PrevBlockHash struct {
		Data string `json:"data"`
	} `json:"prevBlockHash"`
	Type         int `json:"type"`
	Transactions []struct {
		TimeStamp int    `json:"timeStamp"`
		Amount    int    `json:"amount"`
		Signature string `json:"signature"`
		Fee       int    `json:"fee"`
		Recipient string `json:"recipient"`
		Type      int    `json:"type"`
		Deadline  int    `json:"deadline"`
		Message   struct {
		} `json:"message"`
		Version int    `json:"version"`
		Signer  string `json:"signer"`
	} `json:"transactions"`
	Version int    `json:"version"`
	Signer  string `json:"signer"`
	Height  int64  `json:"height"`
}

type NemAccountResponse struct {
	Meta struct {
		Cosignatories []interface{} `json:"cosignatories"`
		CosignatoryOf []interface{} `json:"cosignatoryOf"`
		Status        string        `json:"status"`
		RemoteStatus  string        `json:"remoteStatus"`
	} `json:"meta"`
	Account struct {
		Address         string      `json:"address"`
		HarvestedBlocks int         `json:"harvestedBlocks"`
		Balance         int64       `json:"balance"`
		Importance      float64     `json:"importance"`
		VestedBalance   int64       `json:"vestedBalance"`
		PublicKey       string      `json:"publicKey"`
		Label           interface{} `json:"label"`
		MultisigInfo    struct {
		} `json:"multisigInfo"`
	} `json:"account"`
}

type NemTXResponse struct {
	Meta struct {
		InnerHash struct {
		} `json:"innerHash"`
		ID   int `json:"id"`
		Hash struct {
			Data string `json:"data"`
		} `json:"hash"`
		Height int64 `json:"height"`
	} `json:"meta"`
	Transaction struct {
		TimeStamp int    `json:"timeStamp"`
		Amount    int    `json:"amount"`
		Signature string `json:"signature"`
		Fee       int    `json:"fee"`
		Recipient string `json:"recipient"`
		Type      int    `json:"type"`
		Deadline  int    `json:"deadline"`
		Message   struct {
		} `json:"message"`
		Version int    `json:"version"`
		Signer  string `json:"signer"`
	} `json:"transaction"`
}

// NemClient is the Nem implementation of the CoinClient
type NemClient struct {
	transport.BaseClient
}

// NewNemClient returns a new client using os variables.
func NewNemClient() (*NemClient, error) {
	u, err := url.Parse(getNodeURL("NEM_URL"))
	if err != nil {
		return nil, err
	}

	return &NemClient{
		BaseClient: transport.BaseClient{
			BaseURL: u,
			Client: &http.Client{
				Timeout: time.Second * 6,
			},
			Log: transport.StdLogger,
		},
	}, nil
}

// GetInfo attempts to get standardised coin info from multiple rpc calls.
func (n NemClient) GetInfo() (*transport.CoinState, error) {
	var info NemGetLastBlockResponse

	if err := n.GET("/chain/last-block", nil, &info); err != nil {
		return nil, err
	}

	return &transport.CoinState{
		Data: transport.CoinData{
			Chain:        "main",
			BlockHeight:  int(info.Height),
			CurrentBlock: info.PrevBlockHash.Data,
		},
	}, nil
}

// GetBalance returns the balance of the address.
func (n NemClient) GetBalance(addr string) (*transport.Balance, error) {
	var account NemAccountResponse

	if err := n.GET("/account/get", map[string]string{"address": addr}, &account); err != nil {
		return nil, err
	}

	return &transport.Balance{
		Data: transport.BalanceData{
			Assets: []transport.Asset{
				{
					Asset:   NemAssetID,
					Balance: fmt.Sprintf("%d", account.Account.Balance),
				},
			},
		},
	}, nil
}

// GetTransactionByHash returns the transaction stored at the given hash.
func (n NemClient) GetTransactionByHash(hash string) (*transport.TransactionResp, error) {
	var tx NemTXResponse

	if err := n.GET("/transaction/get", map[string]string{"hash": hash}, &tx); err != nil {
		return nil, err
	}

	var account NemAccountResponse

	if err := n.GET("/account/get/from-public-key", map[string]string{"publicKey": tx.Transaction.Signer}, &account); err != nil {
		return nil, err
	}

	var info NemGetLastBlockResponse

	if err := n.GET("/chain/last-block", nil, &info); err != nil {
		return nil, err
	}

	included := transport.NewInt64(info.Height - tx.Meta.Height)

	return &transport.TransactionResp{
		Data: struct {
			Transaction transport.Transaction `json:"transaction"`
		}{
			Transaction: transport.Transaction{
				ID:    hash,
				From:  account.Account.Address,
				To:    tx.Transaction.Recipient,
				Value: fmt.Sprintf("%d", tx.Transaction.Amount),
				Confirmations: transport.Confirmations{
					Threshold: transport.ConfirmThresholdValue,
					Confirmed: *included >= *transport.ConfirmThresholdValue,
					Value:     included,
				},
			},
		},
	}, nil
}
