package transport

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"

	"github.com/hugorut/coins-oracle/pkg/transport"
)

var (
	IotaAssetID = "MIOTA"

	iotaAPIURL = "https://api.thetangle.org"
)

// IotaGetInfoResponse represents the JSON returned from a get_info response.
type IotaGetInfoResponse struct {
	Metrics struct {
		Tps1Min          float64 `json:"tps1min"`
		Tps10Min         float64 `json:"tps10min"`
		Ctps10Min        float64 `json:"ctps10min"`
		ConfirmationRate float64 `json:"confirmationRate"`
	} `json:"metrics"`
	Transactions []struct {
		Hash       string `json:"hash"`
		Value      int    `json:"value"`
		ReceivedAt int    `json:"receivedAt"`
	} `json:"transactions"`
	Milestones []struct {
		Index int    `json:"index"`
		Hash  string `json:"hash"`
	} `json:"milestones"`
}

// IotaGetTransactionResponse represents the JSON returned from a transaction response.
type IotaGetTransactionResponse struct {
	Hash                string   `json:"hash"`
	Value               int      `json:"value"`
	Address             string   `json:"address"`
	Timestamp           int      `json:"timestamp"`
	TrunkTransaction    string   `json:"trunkTransaction"`
	BranchTransaction   string   `json:"branchTransaction"`
	Tag                 string   `json:"tag"`
	Bundle              string   `json:"bundle"`
	BundleIndex         int      `json:"bundleIndex"`
	BundleSize          int      `json:"bundleSize"`
	Nonce               string   `json:"nonce"`
	Signature           string   `json:"signature"`
	MilestoneIndex      int      `json:"milestoneIndex"`
	Status              string   `json:"status"`
	BundleNext          string   `json:"bundleNext"`
	References          []string `json:"references"`
	TotalReferences     int      `json:"totalReferences"`
	ConfirmingTimestamp int      `json:"confirmingTimestamp"`
}

// IotaGetAddressResponse represents the JSON returned from a address response.
type IotaGetAddressResponse struct {
	Hash              string `json:"hash"`
	Balance           int    `json:"balance"`
	TotalTransactions int    `json:"totalTransactions"`
	Transactions      []struct {
		Hash      string `json:"hash"`
		Value     int    `json:"value"`
		Address   string `json:"address"`
		Bundle    string `json:"bundle"`
		Timestamp int    `json:"timestamp"`
		Status    string `json:"status"`
	} `json:"transactions"`
	Labels []string `json:"labels"`
}

// IotaGetBundleResponse represents the JSON returned from a bundle response which is used to identify a transaction sender/receiver.
type IotaGetBundleResponse struct {
	Hash        string `json:"hash"`
	Attachments []struct {
		Timestamp int    `json:"timestamp"`
		Status    string `json:"status"`
		Inputs    []struct {
			Hash              string `json:"hash"`
			Value             int    `json:"value"`
			Address           string `json:"address"`
			Timestamp         int    `json:"timestamp"`
			TrunkTransaction  string `json:"trunkTransaction"`
			BranchTransaction string `json:"branchTransaction"`
			Bundle            string `json:"bundle"`
			BundleIndex       int    `json:"bundleIndex"`
			BundleSize        int    `json:"bundleSize"`
			Status            string `json:"status"`
		} `json:"inputs"`
		Outputs []struct {
			Hash              string `json:"hash"`
			Value             int    `json:"value"`
			Address           string `json:"address"`
			Timestamp         int    `json:"timestamp"`
			TrunkTransaction  string `json:"trunkTransaction"`
			BranchTransaction string `json:"branchTransaction"`
			Bundle            string `json:"bundle"`
			BundleIndex       int    `json:"bundleIndex"`
			BundleSize        int    `json:"bundleSize"`
			Status            string `json:"status"`
		} `json:"outputs"`
	} `json:"attachments"`
}

// IotaClient is the Iota implementation of the CoinClient
type IotaClient struct {
	transport.BaseClient
}

// NewIotaClient returns a new client using os variables.
func NewIotaClient() (*IotaClient, error) {
	u, err := url.Parse(iotaAPIURL)
	if err != nil {
		return nil, err
	}

	return &IotaClient{
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
func (b IotaClient) GetInfo() (*transport.CoinState, error) {
	var info IotaGetInfoResponse
	if err := b.GET("/live/history", nil, &info); err != nil {
		return nil, errors.Wrap(err, "error making info request")
	}

	latest := info.Milestones[0]

	return &transport.CoinState{
		Data: transport.CoinData{
			Chain:        "main",
			BlockHeight:  latest.Index,
			CurrentBlock: latest.Hash,
		},
	}, nil
}

// GetBalance returns the balance of the address.
func (b IotaClient) GetBalance(addr string) (*transport.Balance, error) {
	var wallet IotaGetAddressResponse
	if err := b.GET("/addresses/"+addr, nil, &wallet); err != nil {
		return nil, errors.Wrap(err, "error making address request")
	}

	return &transport.Balance{
		Data: transport.BalanceData{
			Assets: []transport.Asset{
				{
					Asset:   IotaAssetID,
					Balance: fmt.Sprintf("%d", wallet.Balance),
				},
			},
		},
	}, nil
}

// GetTransactionByHash returns the transaction stored at the given hash.
func (b IotaClient) GetTransactionByHash(hash string) (*transport.TransactionResp, error) {
	var tx IotaGetTransactionResponse
	if err := b.GET("/transactions/"+hash, nil, &tx); err != nil {
		return nil, errors.Wrap(err, "error making tx request")
	}

	var bundle IotaGetBundleResponse
	if err := b.GET("/bundles/"+tx.Bundle, nil, &bundle); err != nil {
		return nil, errors.Wrap(err, "error making bundle request")
	}

	attachment := bundle.Attachments[0]

	return &transport.TransactionResp{
		Data: struct {
			Transaction transport.Transaction `json:"transaction"`
		}{
			Transaction: transport.Transaction{
				ID:    hash,
				From:  attachment.Inputs[0].Address,
				To:    attachment.Outputs[0].Address,
				Value: fmt.Sprintf("%d", tx.Value),
				Confirmations: transport.Confirmations{
					Confirmed: attachment.Status == "confirmed",
				},
			},
		},
	}, nil
}
