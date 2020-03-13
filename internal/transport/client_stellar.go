package transport

import (
	"github.com/hugorut/coins-oracle/pkg/transport"
	"net/http"

	"github.com/stellar/go/protocols/horizon/operations"

	"github.com/stellar/go/clients/horizonclient"
)

var (
	StellarAssetID = "XLM"
)

// StellarClient is the Stellar implementation of the CoinClient
type StellarClient struct {
	Client *horizonclient.Client
}

// NewStellarClient returns a new client using os variables.
func NewStellarClient() (*StellarClient, error) {
	return &StellarClient{
		Client: &horizonclient.Client{
			HorizonURL: getNodeURL("STELLAR_URL"),
			HTTP: &http.Client{
				Timeout: transport.DefaultClientTimeout,
			},
		},
	}, nil
}

// GetInfo attempts to get standardised coin info from multiple rpc calls.
func (s StellarClient) GetInfo() (*transport.CoinState, error) {
	info, err := s.Client.Ledgers(horizonclient.LedgerRequest{
		Order: "desc",
		Limit: 1,
	})
	if err != nil {
		return nil, err
	}

	return &transport.CoinState{
		Data: transport.CoinData{
			Chain:        "main",
			BlockHeight:  int(info.Embedded.Records[0].Sequence),
			CurrentBlock: info.Embedded.Records[0].Hash,
		},
	}, nil
}

// GetBalance returns the balance of the address.
func (s StellarClient) GetBalance(addr string) (*transport.Balance, error) {
	acc, err := s.Client.AccountDetail(horizonclient.AccountRequest{
		AccountID: addr,
	})
	if err != nil {
		return nil, err
	}

	assets := make([]transport.Asset, len(acc.Balances))
	for key, balance := range acc.Balances {
		code := balance.Asset.Code
		if code == "" {
			code = StellarAssetID
		}

		assets[key] = transport.Asset{
			Asset:   code,
			Balance: balance.Balance,
		}
	}

	return &transport.Balance{
		Data: transport.BalanceData{
			Assets: assets,
		},
	}, nil
}

// GetTransactionByHash returns the transaction stored at the given hash.
func (s StellarClient) GetTransactionByHash(hash string) (*transport.TransactionResp, error) {
	tx, err := s.Client.TransactionDetail(hash)
	if err != nil {
		return nil, err
	}

	ops, err := s.Client.Operations(horizonclient.OperationRequest{
		ForTransaction: tx.ID,
	})
	if err != nil {
		return nil, err
	}

	payment := ops.Embedded.Records[0].(operations.Payment)

	return &transport.TransactionResp{
		Data: struct {
			Transaction transport.Transaction `json:"transaction"`
		}{
			Transaction: transport.Transaction{
				ID:    tx.ID,
				From:  payment.From,
				To:    payment.To,
				Value: payment.Amount,
				Confirmations: transport.Confirmations{
					// if the transaction appears on the ledger it is confirmed
					// see https://stellar.stackexchange.com/questions/1464/is-a-payment-returned-through-horizon-api-call-payments-for-account-always-c
					Confirmed: true,
				},
			},
		},
	}, nil
}
