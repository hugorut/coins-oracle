package transport

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/pkg/errors"

	"github.com/hugorut/coins-oracle/pkg/transport"
)

var (
	RippleAssetID = "XRP"
	dropFactor    = 1000000
)

// RippleGetInfoResponse defines the json response returned from a successful server_info call.
type RippleGetInfoResponse struct {
	Result struct {
		Info struct {
			BuildVersion    string `json:"build_version"`
			CompleteLedgers string `json:"complete_ledgers"`
			Hostid          string `json:"hostid"`
			IoLatencyMs     int    `json:"io_latency_ms"`
			JqTransOverflow string `json:"jq_trans_overflow"`
			LastClose       struct {
				ConvergeTimeS float64 `json:"converge_time_s"`
				Proposers     int     `json:"proposers"`
			} `json:"last_close"`
			LoadFactor               float64 `json:"load_factor"`
			LoadFactorServer         int     `json:"load_factor_server"`
			PeerDisconnects          string  `json:"peer_disconnects"`
			PeerDisconnectsResources string  `json:"peer_disconnects_resources"`
			Peers                    int     `json:"peers"`
			PubkeyNode               string  `json:"pubkey_node"`
			ServerState              string  `json:"server_state"`
			ServerStateDurationUs    string  `json:"server_state_duration_us"`
			StateAccounting          struct {
				Connected struct {
					DurationUs  string `json:"duration_us"`
					Transitions int    `json:"transitions"`
				} `json:"connected"`
				Disconnected struct {
					DurationUs  string `json:"duration_us"`
					Transitions int    `json:"transitions"`
				} `json:"disconnected"`
				Full struct {
					DurationUs  string `json:"duration_us"`
					Transitions int    `json:"transitions"`
				} `json:"full"`
				Syncing struct {
					DurationUs  string `json:"duration_us"`
					Transitions int    `json:"transitions"`
				} `json:"syncing"`
				Tracking struct {
					DurationUs  string `json:"duration_us"`
					Transitions int    `json:"transitions"`
				} `json:"tracking"`
			} `json:"state_accounting"`
			Time             string       `json:"time"`
			Uptime           int          `json:"uptime"`
			ValidatedLedger  RippleLedger `json:"validated_ledger"`
			ClosedLedger     RippleLedger `json:"closed_ledger"`
			ValidationQuorum int          `json:"validation_quorum"`
		} `json:"info"`
		Status string `json:"status"`
	} `json:"result"`
}

type RippleLedger struct {
	Age            int     `json:"age"`
	BaseFeeXrp     float64 `json:"base_fee_xrp"`
	Hash           string  `json:"hash"`
	ReserveBaseXrp int     `json:"reserve_base_xrp"`
	ReserveIncXrp  int     `json:"reserve_inc_xrp"`
	Seq            int     `json:"seq"`
}

// RippleTxResponse defines a struct that represents a successful json response from the tx endpoint.
type RippleTxResponse struct {
	Result struct {
		Account         string          `json:"Account"`
		Amount          json.RawMessage `json:"Amount"`
		Destination     string          `json:"Destination"`
		DestinationTag  int             `json:"DestinationTag"`
		Fee             string          `json:"Fee"`
		Flags           int64           `json:"Flags"`
		Sequence        int             `json:"Sequence"`
		SigningPubKey   string          `json:"SigningPubKey"`
		TransactionType string          `json:"TransactionType"`
		TxnSignature    string          `json:"TxnSignature"`
		Date            int             `json:"date"`
		Hash            string          `json:"hash"`
		InLedger        int             `json:"inLedger"`
		LedgerIndex     int             `json:"ledger_index"`
		Meta            struct {
			AffectedNodes []struct {
				ModifiedNode struct {
					FinalFields struct {
						Account    string `json:"Account"`
						Balance    string `json:"Balance"`
						Flags      int    `json:"Flags"`
						OwnerCount int    `json:"OwnerCount"`
						Sequence   int    `json:"Sequence"`
					} `json:"FinalFields"`
					LedgerEntryType string `json:"LedgerEntryType"`
					LedgerIndex     string `json:"LedgerIndex"`
					PreviousFields  struct {
						Balance string `json:"Balance"`
					} `json:"PreviousFields"`
					PreviousTxnID     string `json:"PreviousTxnID"`
					PreviousTxnLgrSeq int    `json:"PreviousTxnLgrSeq"`
				} `json:"ModifiedNode"`
			} `json:"AffectedNodes"`
			TransactionIndex  int    `json:"TransactionIndex"`
			TransactionResult string `json:"TransactionResult"`
			DeliveredAmount   string `json:"delivered_amount"`
		} `json:"meta"`
		Status    string `json:"status"`
		Validated bool   `json:"validated"`
	} `json:"result"`
}

// RippleComplexAmount represents a complex description of an transaction amount.
type RippleComplexAmount struct {
	Currency string `json:"currency"`
	Issuer   string `json:"issuer"`
	Value    string `json:"value"`
}

// RippleAccountInfoResponse defines a struct that represents the succesful json response from a successful account_info call.
type RippleAccountInfoResponse struct {
	Result struct {
		AccountData struct {
			Account           string `json:"Account"`
			Balance           string `json:"Balance"`
			Flags             int    `json:"Flags"`
			LedgerEntryType   string `json:"LedgerEntryType"`
			OwnerCount        int    `json:"OwnerCount"`
			PreviousTxnID     string `json:"PreviousTxnID"`
			PreviousTxnLgrSeq int    `json:"PreviousTxnLgrSeq"`
			Sequence          int    `json:"Sequence"`
			Index             string `json:"index"`
		} `json:"account_data"`
		LedgerCurrentIndex int    `json:"ledger_current_index"`
		Status             string `json:"status"`
		Validated          bool   `json:"validated"`
	} `json:"result"`
}

// RippleRPCRequest defines a common rpc request json.
type RippleRPCRequest struct {
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
}

// RippleGetBalanceParams defines a struct that represents the json to be used under the
// rpc params in a account_info call.
type RippleGetBalanceParams struct {
	Account string `json:"account"`
}

// RippleTxParams defines a struct that represents the json to be used under the
// rpc params in a tx call.
type RippleTxParams struct {
	Transaction string `json:"transaction"`
}

// RippleClient is the Ripple implementation of the CoinClient
type RippleClient struct {
	transport.BaseClient
}

// NewRippleClient returns a new client using os variables.
func NewRippleClient() (*RippleClient, error) {
	u, err := url.Parse(getNodeURL("RIPPLE_URL"))
	if err != nil {
		return nil, err
	}

	return &RippleClient{
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
func (rc RippleClient) GetInfo() (*transport.CoinState, error) {
	var info RippleGetInfoResponse

	err := rc.POST(&RippleRPCRequest{
		Method: "server_info",
	}, "/", &info)
	if err != nil {
		return nil, err
	}

	hash := info.Result.Info.ValidatedLedger.Hash
	count := info.Result.Info.ValidatedLedger.Seq

	if hash == "" {
		hash = info.Result.Info.ClosedLedger.Hash
		count = info.Result.Info.ClosedLedger.Seq
	}

	return &transport.CoinState{
		Data: transport.CoinData{
			Chain:        "main",
			BlockHeight:  count,
			CurrentBlock: hash,
		},
	}, nil
}

// GetBalance returns the balance of the address.
func (rc RippleClient) GetBalance(addr string) (*transport.Balance, error) {
	var info RippleAccountInfoResponse

	err := rc.POST(&RippleRPCRequest{
		Method: "account_info",
		Params: []interface{}{
			RippleGetBalanceParams{
				Account: addr,
			},
		},
	}, "/", &info)
	if err != nil {
		return nil, err
	}

	i, err := strconv.ParseInt(info.Result.AccountData.Balance, 10, 64)
	if err != nil {
		return nil, err
	}

	value := strconv.FormatFloat(float64(i)/float64(dropFactor), 'f', -1, 64)

	return &transport.Balance{
		Data: transport.BalanceData{
			Assets: []transport.Asset{
				{
					Asset:   RippleAssetID,
					Balance: value,
				},
			},
		},
	}, nil
}

// GetTransactionByHash returns the transaction stored at the given hash.
func (rc RippleClient) GetTransactionByHash(hash string) (*transport.TransactionResp, error) {
	var info RippleTxResponse

	err := rc.POST(&RippleRPCRequest{
		Method: "tx",
		Params: []interface{}{
			RippleTxParams{
				Transaction: hash,
			},
		},
	}, "/", &info)
	if err != nil {
		return nil, err
	}

	value, err := getTransactionValue(info.Result.Amount)
	if err != nil {
		return nil, errors.Wrap(err, "error getting ripple transaction value from raw messag")
	}

	return &transport.TransactionResp{
		Data: struct {
			Transaction transport.Transaction `json:"transaction"`
		}{
			Transaction: transport.Transaction{
				ID:    info.Result.Hash,
				From:  info.Result.Account,
				To:    info.Result.Destination,
				Value: value,
				Confirmations: transport.Confirmations{
					Confirmed: info.Result.Validated,
				},
			},
		},
	}, nil
}

func getTransactionValue(amount json.RawMessage) (string, error) {
	var res RippleComplexAmount
	err := json.Unmarshal(amount, &res)
	if err == nil {
		i, err := strconv.ParseInt(res.Value, 10, 64)
		if err != nil {
			return "", err
		}

		return format(i), nil
	}

	var str string
	err = json.Unmarshal(amount, &str)
	if err != nil {
		return "", err
	}

	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return "", err
	}

	return format(i), nil
}

func format(i int64) string {
	return strconv.FormatFloat(float64(i)/float64(dropFactor), 'f', -1, 64)
}
