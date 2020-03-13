package transport

import (
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"

	"github.com/hugorut/coins-oracle/pkg/transport"
)

var (
	TezosAssetID = "XTZ"

	tezosAPIURL = "https://api.tezos.id"
)

// TezosBlockResponse is a struct representing a blocks JSON response.
type TezosBlockResponse struct {
	Protocol string `json:"protocol"`
	ChainID  string `json:"chain_id"`
	Hash     string `json:"hash"`
	Header   struct {
		Level            int       `json:"level"`
		Proto            int       `json:"proto"`
		Predecessor      string    `json:"predecessor"`
		Timestamp        time.Time `json:"timestamp"`
		ValidationPass   int       `json:"validation_pass"`
		OperationsHash   string    `json:"operations_hash"`
		Fitness          []string  `json:"fitness"`
		Context          string    `json:"context"`
		Priority         int       `json:"priority"`
		ProofOfWorkNonce string    `json:"proof_of_work_nonce"`
		Signature        string    `json:"signature"`
	} `json:"header"`
	Metadata struct {
		Protocol        string `json:"protocol"`
		NextProtocol    string `json:"next_protocol"`
		TestChainStatus struct {
			Status string `json:"status"`
		} `json:"test_chain_status"`
		MaxOperationsTTL       int `json:"max_operations_ttl"`
		MaxOperationDataLength int `json:"max_operation_data_length"`
		MaxBlockHeaderLength   int `json:"max_block_header_length"`
		MaxOperationListLength []struct {
			MaxSize int `json:"max_size"`
			MaxOp   int `json:"max_op,omitempty"`
		} `json:"max_operation_list_length"`
		Baker string `json:"baker"`
		Level struct {
			Level                int  `json:"level"`
			LevelPosition        int  `json:"level_position"`
			Cycle                int  `json:"cycle"`
			CyclePosition        int  `json:"cycle_position"`
			VotingPeriod         int  `json:"voting_period"`
			VotingPeriodPosition int  `json:"voting_period_position"`
			ExpectedCommitment   bool `json:"expected_commitment"`
		} `json:"level"`
		VotingPeriodKind string        `json:"voting_period_kind"`
		NonceHash        interface{}   `json:"nonce_hash"`
		ConsumedGas      string        `json:"consumed_gas"`
		Deactivated      []interface{} `json:"deactivated"`
		BalanceUpdates   []struct {
			Kind     string `json:"kind"`
			Contract string `json:"contract,omitempty"`
			Change   string `json:"change"`
			Category string `json:"category,omitempty"`
			Delegate string `json:"delegate,omitempty"`
			Level    int    `json:"level,omitempty"`
		} `json:"balance_updates"`
	} `json:"metadata"`
	Operations [][]struct {
		Protocol string `json:"protocol"`
		ChainID  string `json:"chain_id"`
		Hash     string `json:"hash"`
		Branch   string `json:"branch"`
		Contents []struct {
			Kind     string `json:"kind"`
			Level    int    `json:"level"`
			Metadata struct {
				BalanceUpdates []struct {
					Kind     string `json:"kind"`
					Contract string `json:"contract,omitempty"`
					Change   string `json:"change"`
					Category string `json:"category,omitempty"`
					Delegate string `json:"delegate,omitempty"`
					Level    int    `json:"level,omitempty"`
				} `json:"balance_updates"`
				Delegate string `json:"delegate"`
				Slots    []int  `json:"slots"`
			} `json:"metadata"`
		} `json:"contents"`
		Signature string `json:"signature"`
	} `json:"operations"`
}

// TezosGetBalanceResponse is a struct representing a balance JSON response.
type TezosGetBalanceResponse struct {
	Manager   string `json:"manager"`
	Balance   string `json:"balance"`
	Spendable bool   `json:"spendable"`
	Delegate  struct {
		Setable bool   `json:"setable"`
		Value   string `json:"value"`
	} `json:"delegate"`
	Counter string `json:"counter"`
}

// TezosGetTransactionResponse is a struct representing a tx JSON response.
type TezosGetTransactionResponse []struct {
	Tx struct {
		StorageLimit                                string      `json:"storageLimit"`
		Destination                                 string      `json:"destination"`
		Amount                                      string      `json:"amount"`
		OpUUID                                      string      `json:"opUuid"`
		OperationResultStorage                      interface{} `json:"operationResultStorage"`
		UUID                                        string      `json:"uuid"`
		GasLimit                                    string      `json:"gasLimit"`
		Kind                                        string      `json:"kind"`
		OperationResultStorageSize                  string      `json:"operationResultStorageSize"`
		OperationResultStatus                       string      `json:"operationResultStatus"`
		BlockHash                                   string      `json:"blockHash"`
		OperationResultAllocatedDestinationContract interface{} `json:"operationResultAllocatedDestinationContract"`
		Fee                                         string      `json:"fee"`
		OperationResultUUID                         string      `json:"operationResultUuid"`
		OperationResultConsumedGas                  string      `json:"operationResultConsumedGas"`
		OperationResultBigMapDiff                   interface{} `json:"operationResultBigMapDiff"`
		Counter                                     string      `json:"counter"`
		BlockLevel                                  int         `json:"blockLevel"`
		OperationResultErrors                       interface{} `json:"operationResultErrors"`
		BlockTimestamp                              time.Time   `json:"blockTimestamp"`
		Parameters                                  interface{} `json:"parameters"`
		Source                                      string      `json:"source"`
		InsertedTimestamp                           string      `json:"insertedTimestamp"`
		MetadataUUID                                string      `json:"metadataUuid"`
		OperationResultPaidStorageSizeDiff          interface{} `json:"operationResultPaidStorageSizeDiff"`
	} `json:"tx"`
	Op struct {
		Signature         string    `json:"signature"`
		BlockUUID         string    `json:"blockUuid"`
		OpHash            string    `json:"opHash"`
		UUID              string    `json:"uuid"`
		ChainID           string    `json:"chainId"`
		BlockHash         string    `json:"blockHash"`
		Protocol          string    `json:"protocol"`
		Branch            string    `json:"branch"`
		BlockLevel        int       `json:"blockLevel"`
		BlockTimestamp    time.Time `json:"blockTimestamp"`
		InsertedTimestamp string    `json:"insertedTimestamp"`
	} `json:"op"`
}

// TezosClient is the Tezos implementation of the CoinClient
type TezosClient struct {
	transport.BaseClient
	APIClient transport.BaseClient
}

// NewTezosClient returns a new client using os variables.
func NewTezosClient() (*TezosClient, error) {
	u, err := url.Parse(getNodeURL("TEZOS_URL"))
	if err != nil {
		return nil, err
	}

	tu, _ := url.Parse(tezosAPIURL)

	return &TezosClient{
		BaseClient: transport.BaseClient{
			BaseURL: u,
			Client: &http.Client{
				Timeout: transport.DefaultClientTimeout,
			},
			Log: transport.StdLogger,
		},
		APIClient: transport.BaseClient{
			BaseURL: tu,
			Client: &http.Client{
				Timeout: transport.DefaultClientTimeout,
			},
			Log: transport.StdLogger,
		},
	}, nil
}

// GetInfo attempts to get standardised coin info from multiple rpc calls.
func (b TezosClient) GetInfo() (*transport.CoinState, error) {
	var blocks TezosBlockResponse

	if err := b.GET("/chains/main/blocks/head", nil, &blocks); err != nil {
		return nil, errors.Wrap(err, "error getting tezos blocks")
	}

	return &transport.CoinState{
		Data: transport.CoinData{
			Chain:        "main",
			BlockHeight:  blocks.Header.Level,
			CurrentBlock: blocks.Operations[0][0].Hash,
		},
	}, nil
}

// GetBalance returns the balance of the address.
func (b TezosClient) GetBalance(addr string) (*transport.Balance, error) {
	var balance TezosGetBalanceResponse

	if err := b.GET("/chains/main/balance/head/context/contracts/"+addr, nil, &balance); err != nil {
		return nil, errors.Wrap(err, "error getting tezos balance")
	}

	return &transport.Balance{
		Data: transport.BalanceData{
			Assets: []transport.Asset{
				{
					Asset:   TezosAssetID,
					Balance: balance.Balance,
				},
			},
		},
	}, nil
}

// GetTransactionByHash returns the transaction stored at the given hash.
func (b TezosClient) GetTransactionByHash(hash string) (*transport.TransactionResp, error) {
	var txs TezosGetTransactionResponse

	if err := b.APIClient.GET("/mooncake/mainnet/v1/transactions", map[string]string{"op": hash}, &txs); err != nil {
		return nil, errors.Wrap(err, "error getting tezos tx")
	}

	tx := txs[0]

	return &transport.TransactionResp{
		Data: struct {
			Transaction transport.Transaction `json:"transaction"`
		}{
			Transaction: transport.Transaction{
				ID:    hash,
				From:  tx.Tx.Source,
				To:    tx.Tx.Destination,
				Value: tx.Tx.Amount,
				Confirmations: transport.Confirmations{
					Confirmed: true,
				},
			},
		},
	}, nil
}
