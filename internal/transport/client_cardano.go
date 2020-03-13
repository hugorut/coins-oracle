package transport

import (
	"encoding/json"
	"errors"
	"github.com/hugorut/coins-oracle/pkg/transport"
	"net/http"
	"net/url"
)

var (
	CardanoAssetID = "ADA"
)

// CardanoBaseResponse defines a struct which represents the cardano base json message
type CardanoBaseResponse struct {
	Right []json.RawMessage `json:"Right"`
}

// CardanoBlock a json response for a cardano block
type CardanoBlock struct {
	CbeEpoch      int    `json:"cbeEpoch"`
	CbeSlot       int    `json:"cbeSlot"`
	CbeBlkHash    string `json:"cbeBlkHash"`
	CbeTimeIssued int    `json:"cbeTimeIssued"`
	CbeTxNum      int    `json:"cbeTxNum"`
	CbeTotalSent  struct {
		GetCoin string `json:"getCoin"`
	} `json:"cbeTotalSent"`
	CbeSize      int    `json:"cbeSize"`
	CbeBlockLead string `json:"cbeBlockLead"`
	CbeFees      struct {
		GetCoin string `json:"getCoin"`
	} `json:"cbeFees"`
}

// CardanoAccountResponse represents a successful json response for get account details.
type CardanoAccountResponse struct {
	Right struct {
		CaAddress string `json:"caAddress"`
		CaType    string `json:"caType"`
		CaTxNum   int    `json:"caTxNum"`
		CaBalance struct {
			GetCoin string `json:"getCoin"`
		} `json:"caBalance"`
		CaTxList []struct {
			CtbID         string              `json:"ctbId"`
			CtbTimeIssued int                 `json:"ctbTimeIssued"`
			CtbInputs     [][]json.RawMessage `json:"ctbInputs"`
			CtbOutputs    [][]json.RawMessage `json:"ctbOutputs"`
			CtbInputSum   struct {
				GetCoin string `json:"getCoin"`
			} `json:"ctbInputSum"`
			CtbOutputSum struct {
				GetCoin string `json:"getCoin"`
			} `json:"ctbOutputSum"`
		} `json:"caTxList"`
	} `json:"Right"`
}

// CardanoGetTransactionResponse represents a successful json response returned from a cardano
type CardanoGetTransactionResponse struct {
	Right struct {
		CtsID              string      `json:"ctsId"`
		CtsTxTimeIssued    int         `json:"ctsTxTimeIssued"`
		CtsBlockTimeIssued int         `json:"ctsBlockTimeIssued"`
		CtsBlockHeight     int         `json:"ctsBlockHeight"`
		CtsBlockEpoch      int         `json:"ctsBlockEpoch"`
		CtsBlockSlot       int         `json:"ctsBlockSlot"`
		CtsBlockHash       string      `json:"ctsBlockHash"`
		CtsRelayedBy       interface{} `json:"ctsRelayedBy"`
		CtsTotalInput      struct {
			GetCoin string `json:"getCoin"`
		} `json:"ctsTotalInput"`
		CtsTotalOutput struct {
			GetCoin string `json:"getCoin"`
		} `json:"ctsTotalOutput"`
		CtsFees struct {
			GetCoin string `json:"getCoin"`
		} `json:"ctsFees"`
		CtsInputs  [][]json.RawMessage `json:"ctsInputs"`
		CtsOutputs [][]json.RawMessage `json:"ctsOutputs"`
	} `json:"Right"`
}

// GetCoin represents a getcoin json structure
type GetCoin struct {
	GetCoin string `json:"getCoin"`
}

// CardanoTXHeader holds information of a transfer of funds.
type CardanoTXHeader struct {
	// Account is who the transfer is from or to
	Account string
	// Value is the the value of the transfer
	Value string
}

// CardanoClient is the Cardano implementation of the CoinClient
type CardanoClient struct {
	transport.BaseClient
}

// NewCardanoClient returns a new client using os variables.
func NewCardanoClient() (*CardanoClient, error) {
	u, err := url.Parse(getNodeURL("CARDANO_URL"))
	if err != nil {
		return nil, err
	}

	return &CardanoClient{
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
func (c CardanoClient) GetInfo() (*transport.CoinState, error) {
	block, err := c.getInfo()
	if err != nil {
		return nil, err
	}

	return &transport.CoinState{
		Data: transport.CoinData{
			Chain:        "main",
			BlockHeight:  block.CbeSlot,
			CurrentBlock: block.CbeBlkHash,
		},
	}, nil
}

// GetBalance returns the balance of the address.
func (c CardanoClient) GetBalance(addr string) (*transport.Balance, error) {
	var account CardanoAccountResponse

	err := c.GET("/api/addresses/summary/"+addr, nil, &account)
	if err != nil {
		return nil, err
	}

	return &transport.Balance{
		Data: transport.BalanceData{
			Assets: []transport.Asset{
				{
					Asset:   CardanoAssetID,
					Balance: account.Right.CaBalance.GetCoin,
				},
			},
		},
	}, nil
}

// GetTransactionByHash returns the transaction stored at the given hash.
func (c CardanoClient) GetTransactionByHash(hash string) (*transport.TransactionResp, error) {
	var transaction CardanoGetTransactionResponse

	err := c.GET("/api/txs/summary/"+hash, nil, &transaction)
	if err != nil {
		return nil, err
	}

	fromHeader, err := getTransactionHeader(transaction.Right.CtsInputs[0])
	if err != nil {
		return nil, err
	}

	toHeader, err := getTransactionHeader(transaction.Right.CtsOutputs[0])
	if err != nil {
		return nil, err
	}

	return &transport.TransactionResp{
		Data: struct {
			Transaction transport.Transaction `json:"transaction"`
		}{
			Transaction: transport.Transaction{
				ID:    transaction.Right.CtsID,
				From:  fromHeader.Account,
				To:    toHeader.Account,
				Value: fromHeader.Value,
				Confirmations: transport.Confirmations{
					Confirmed: true,
				},
			},
		},
	}, nil
}

func (c CardanoClient) getInfo() (*CardanoBlock, error) {
	var base CardanoBaseResponse

	err := c.GET("/api/blocks/pages", nil, &base)
	if err != nil {
		return nil, err
	}

	var blocks []CardanoBlock
	for _, item := range base.Right {
		err := json.Unmarshal(item, &blocks)
		if err != nil {
			continue
		}
		break
	}

	if len(blocks) == 0 {
		return nil, errors.New("no cardano blocks returned")
	}

	return &blocks[0], nil
}

func getTransactionHeader(input []json.RawMessage) (*CardanoTXHeader, error) {
	var acc string
	var info GetCoin
	for _, value := range input {
		if err := json.Unmarshal(value, &acc); err == nil {
			continue
		}

		if err := json.Unmarshal(value, &info); err == nil {
			break
		}
	}

	if acc == "" || info.GetCoin == "" {
		return nil, errors.New("could not cardano transaction header")
	}

	return &CardanoTXHeader{
		Account: acc,
		Value:   info.GetCoin,
	}, nil
}
