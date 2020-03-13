package transport

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/btcsuite/btcutil/base58"

	"github.com/hugorut/coins-oracle/pkg/transport"
)

var (
	TronAssetID = "TRX"
)

// TronGetTXResponse represents a successful json response from the gettransactionbyid endpoint.
type TronGetTXResponse struct {
	Signature []string `json:"signature"`
	TxID      string   `json:"txID"`
	RawData   struct {
		Contract []struct {
			Parameter struct {
				Value struct {
					Amount       int    `json:"amount"`
					AssetName    string `json:"asset_name"`
					OwnerAddress string `json:"owner_address"`
					ToAddress    string `json:"to_address"`
				} `json:"value"`
				TypeURL string `json:"type_url"`
			} `json:"parameter"`
			Type string `json:"type"`
		} `json:"contract"`
		RefBlockBytes string `json:"ref_block_bytes"`
		RefBlockHash  string `json:"ref_block_hash"`
		Expiration    int64  `json:"expiration"`
		Timestamp     int64  `json:"timestamp"`
	} `json:"raw_data"`
	RawDataHex string `json:"raw_data_hex"`
}

// TronGetTXInfoResponse represents a successful json response from the gettransactioninfobyid endpoint.
type TronGetTXInfoResponse struct {
	ID             string   `json:"id"`
	Fee            int      `json:"fee"`
	BlockNumber    int      `json:"blockNumber"`
	BlockTimeStamp int64    `json:"blockTimeStamp"`
	ContractResult []string `json:"contractResult"`
	Receipt        struct {
		NetFee int `json:"net_fee"`
	} `json:"receipt"`
}

// TronGetInfoResponse represents a successful json response from the getnowblock endpoint.
type TronGetInfoResponse struct {
	BlockID     string `json:"blockID"`
	BlockHeader struct {
		RawData struct {
			Number         int    `json:"number"`
			TxTrieRoot     string `json:"txTrieRoot"`
			WitnessAddress string `json:"witness_address"`
			ParentHash     string `json:"parentHash"`
			Timestamp      int64  `json:"timestamp"`
		} `json:"raw_data"`
		WitnessSignature string `json:"witness_signature"`
	} `json:"block_header"`
}

// TronGetBalanceResponse represents a successful json response getaccount endpoint.
type TronGetBalanceResponse struct {
	AccountName string `json:"account_name"`
	Address     string `json:"address"`
	Balance     int    `json:"balance"`
	Asset       []struct {
		Key   string `json:"key"`
		Value int    `json:"value"`
	} `json:"asset"`
	Frozen []struct {
		FrozenBalance int   `json:"frozen_balance"`
		ExpireTime    int64 `json:"expire_time"`
	} `json:"frozen"`
	CreateTime         int64 `json:"create_time"`
	LatestOprationTime int64 `json:"latest_opration_time"`
	FreeNetUsage       int   `json:"free_net_usage"`
	FreeAssetNetUsage  []struct {
		Key   string `json:"key"`
		Value int    `json:"value"`
	} `json:"free_asset_net_usage"`
	LatestConsumeTime     int64 `json:"latest_consume_time"`
	LatestConsumeFreeTime int64 `json:"latest_consume_free_time"`
	AccountResource       struct {
		LatestConsumeTimeForEnergy int64 `json:"latest_consume_time_for_energy"`
	} `json:"account_resource"`
	AssetV2 []struct {
		Key   string `json:"key"`
		Value int    `json:"value"`
	} `json:"assetV2"`
	FreeAssetNetUsageV2 []struct {
		Key   string `json:"key"`
		Value int    `json:"value"`
	} `json:"free_asset_net_usageV2"`
}

// TronGetAddressReq represents a json body for the getaddress.
type TronGetAddressReq struct {
	Address string `json:"address"`
}

// TronGetTXReq represents a json body for the gettransactionbyid.
type TronGetTXReq struct {
	Value string `json:"value"`
}

// TronClient is the Tron implementation of the CoinClient
type TronClient struct {
	transport.BaseClient
}

// NewTronClient returns a new client using os variables.
func NewTronClient() (*TronClient, error) {
	u, err := url.Parse(getNodeURL("TRON_URL"))
	if err != nil {
		return nil, err
	}

	return &TronClient{
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
func (t TronClient) GetInfo() (*transport.CoinState, error) {
	var info TronGetInfoResponse

	if err := t.POST(nil, "/wallet/getnowblock", &info); err != nil {
		return nil, err
	}

	return &transport.CoinState{
		Data: transport.CoinData{
			Chain:        "main",
			BlockHeight:  info.BlockHeader.RawData.Number,
			CurrentBlock: info.BlockID,
		},
	}, nil
}

// GetBalance returns the balance of the address.
func (t TronClient) GetBalance(addr string) (*transport.Balance, error) {
	var acc TronGetBalanceResponse

	if err := t.POST(TronGetAddressReq{Address: base58ToHex(addr)}, "/wallet/getaccount", &acc); err != nil {
		return nil, err
	}

	return &transport.Balance{
		Data: transport.BalanceData{
			Assets: []transport.Asset{
				{
					Asset:   TronAssetID,
					Balance: fmt.Sprintf("%d", acc.Balance),
				},
			},
		},
	}, nil
}

// GetTransactionByHash returns the transaction stored at the given hash.
func (t TronClient) GetTransactionByHash(hash string) (*transport.TransactionResp, error) {
	var tx TronGetTXResponse
	if err := t.POST(TronGetTXReq{Value: hash}, "/wallet/gettransactionbyid", &tx); err != nil {
		return nil, err
	}

	var info TronGetTXInfoResponse
	if err := t.POST(TronGetTXReq{Value: hash}, "/wallet/gettransactioninfobyid", &info); err != nil {
		return nil, err
	}

	var latest TronGetInfoResponse
	if err := t.POST(nil, "/wallet/getnowblock", &latest); err != nil {
		return nil, err
	}

	txData := tx.RawData.Contract[0].Parameter.Value
	included := transport.NewInt64(int64(latest.BlockHeader.RawData.Number) - int64(info.BlockNumber))

	return &transport.TransactionResp{
		Data: struct {
			Transaction transport.Transaction `json:"transaction"`
		}{
			Transaction: transport.Transaction{
				ID:    tx.TxID,
				From:  hexToBase58(txData.OwnerAddress),
				To:    hexToBase58(txData.ToAddress),
				Value: fmt.Sprintf("%d", txData.Amount),
				Confirmations: transport.Confirmations{
					Threshold: transport.ConfirmThresholdValue,
					Confirmed: *included >= *transport.ConfirmThresholdValue,
					Value:     included,
				},
			},
		},
	}, nil
}

func base58ToHex(input string) string {
	num := base58.Decode(input)

	str := hex.EncodeToString(num)

	// return removing checksum (8 char at end)
	return strings.ToUpper(str[0 : len(str)-8])
}

func hexToBase58(input string) string {
	raw, _ := hex.DecodeString(input)

	hash0 := sha256.Sum256(raw)
	hash1 := sha256.Sum256(toBytes(hash0))
	checkSum := hash1[0:4]

	var sum []byte
	sum = append(sum, raw...)
	sum = append(sum, checkSum...)

	return base58.Encode(sum)
}

func toBytes(bits [32]byte) []byte {
	data := make([]byte, len(bits))
	for key, value := range bits {
		data[key] = value
	}
	return data
}
