package transport

import (
	"fmt"

	"github.com/eoscanada/eos-go"

	"github.com/hugorut/coins-oracle/pkg/transport"
)

var (
	EosAssetID = "EOS"
)

// EosClient is the Eos implementation of the CoinClient
type EosClient struct {
	Client *eos.API
}

// NewEosClient returns a new client using os variables.
func NewEosClient() (*EosClient, error) {
	api := eos.New(getNodeURL("EOS_URL"))

	return &EosClient{
		Client: api,
	}, nil
}

// GetInfo attempts to get standardised coin info from multiple rpc calls.
func (e EosClient) GetInfo() (*transport.CoinState, error) {
	info, err := e.Client.GetInfo()
	if err != nil {
		return nil, err
	}

	return &transport.CoinState{
		Data: transport.CoinData{
			Chain:        info.ChainID.String(),
			BlockHeight:  int(info.HeadBlockNum),
			CurrentBlock: info.HeadBlockID.String(),
		},
	}, nil
}

// GetBalance returns the balance of the address.
func (e EosClient) GetBalance(addr string) (*transport.Balance, error) {
	name := eos.AccountName(addr)
	code := eos.AccountName("eosio.token")

	balance, err := e.Client.GetCurrencyBalance(name, "", code)
	if err != nil {
		return nil, err
	}

	assets := make([]transport.Asset, len(balance))
	for key, a := range balance {
		assets[key] = transport.Asset{
			Asset:   a.Symbol.Symbol,
			Balance: fmt.Sprintf("%d", a.Amount),
		}
	}

	return &transport.Balance{
		Data: transport.BalanceData{
			Assets: assets,
		},
	}, nil
}

// GetTransactionByHash returns the transaction stored at the given hash.
func (e EosClient) GetTransactionByHash(hash string) (*transport.TransactionResp, error) {
	t, err := e.Client.GetTransaction(hash)
	if err != nil {
		return nil, err
	}

	var from string
	var to string
	var value string
	for _, a := range t.Transaction.Transaction.Actions {
		if a.Name == eos.ActionName("transfer") {
			m := a.Data.(map[string]interface{})

			from = fmt.Sprintf("%s", m["from"])
			to = fmt.Sprintf("%s", m["to"])
			value = fmt.Sprintf("%s", m["quantity"])
			break
		}
	}

	info, err := e.Client.GetInfo()
	if err != nil {
		return nil, err
	}

	included := transport.NewInt64(int64(info.HeadBlockNum) - int64(t.BlockNum))

	return &transport.TransactionResp{
		Data: struct {
			Transaction transport.Transaction `json:"transaction"`
		}{
			Transaction: transport.Transaction{
				From:  from,
				ID:    t.ID.String(),
				To:    to,
				Value: value,
				Confirmations: transport.Confirmations{
					Threshold: transport.ConfirmThresholdValue,
					Confirmed: *transport.ConfirmThresholdValue <= *included,
					Value:     included,
				},
			},
		},
	}, nil
}
