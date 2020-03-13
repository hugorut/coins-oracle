package transport

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"time"

	"github.com/pkg/errors"
)

const (
	DefaultClientTimeout = time.Second * 3
	RPCVersion           = "2.0"
)

var (
	ConfirmThresholdValue *int64 = NewInt64(5)
	StdLogger                    = log.New(os.Stderr, "", log.LstdFlags)

	hexReg = regexp.MustCompile("^0x")
)

// NewInt64 returns a new pointer to an int64.
func NewInt64(i int64) *int64 {
	return &i
}

// CoinState contains metadata associated with a specific coin node.
type CoinState struct {
	Data CoinData `json:"data"`
}

// CoinData holds a standardised format for displaying coin blockchain info.
type CoinData struct {
	Chain        string `json:"chain"`
	BlockHeight  int    `json:"block_height"`
	CurrentBlock string `json:"current_block_hash"`
}

// Balance contains information about the current state of a blockchain wallet.
type Balance struct {
	Data BalanceData `json:"data"`
}

// BalanceData holds the main json information for a Balance.
type BalanceData struct {
	// Assets is a slice of assets as some coins wallets can hold multiple balances.
	Assets []Asset `json:"assets"`
}

// Asset holds information about a specific coin balance.
type Asset struct {
	Asset   string `json:"asset"`
	Balance string `json:"balance"`
}

// Transaction represents a specific blockchain transaction.
type Transaction struct {
	ID            string        `json:"id"`
	From          string        `json:"from"`
	To            string        `json:"to"`
	Value         string        `json:"value"`
	Confirmations Confirmations `json:"confirmations"`
}

// Confirmations is a struct to hold the transaction confirmations data
type Confirmations struct {
	Threshold *int64 `json:"threshold,omitempty"`
	Confirmed bool   `json:"confirmed"`
	Value     *int64 `json:"value,omitempty"`
}

// TransactionResp wraps a transaction in a json.api defined response.
type TransactionResp struct {
	Data struct {
		Transaction Transaction `json:"transaction"`
	} `json:"data"`
}

// TransactionsResp wraps a list of transactions in a json.api defined response.
type TransactionsResp struct {
	Data struct {
		Transactions []Transaction `json:"transactions"`
	} `json:"data"`
}

// CoinClient defines an interface that communicates
// with a coin specific lambda function.
type CoinClient interface {
	// GetInfo fetches info on the available node.
	GetInfo() (*CoinState, error)
	// GetBalance fetches the current balance of assets in the address.
	GetBalance(addr string) (*Balance, error)
	// GetTransactionByHash fetches information about a transaction on a ledger by its hash.
	GetTransactionByHash(hash string) (*TransactionResp, error)
}

// AddressImporter defines an interface that a coin client can adhear to.
// If a CoinClient has this interface then it can import and address to watch.
type AddressImporter interface {
	// ImportAddress adds an address to track. Note this will kick of a reindex if the
	// Coin supports such functionality.
	ImportAddress(addr string) error
}

// BaseClient handles some of the more repetitive http client handling
type BaseClient struct {
	BaseURL *url.URL
	Client  *http.Client
	Log     *log.Logger
}

func (b BaseClient) Logf(format string, v ...interface{}) {
	if b.Log == nil {
		return
	}

	b.Log.Printf(format, v...)
}

// GET executes a GET request using the path and query params, marshalling the output to out.
func (b BaseClient) GET(path string, queryP map[string]string, out interface{}) error {
	u := *b.BaseURL
	u.Path = path

	if queryP != nil {
		v := u.Query()

		for key, value := range queryP {
			v.Add(key, value)
		}

		u.RawQuery = v.Encode()
	}

	endpoint := u.String()
	b.Logf("making GET request to %s", endpoint)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}

	res, err := b.Client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	raw, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return errors.Wrap(err, "error reading response body")
	}

	b.Logf("received response: %s from GET to: %s", string(raw), endpoint)

	if reflect.TypeOf(out).Kind() == reflect.String {
		out = string(raw)
		return err
	}

	return json.Unmarshal(raw, out)
}

// POST executes a POST request using the path and body, marshalling the output to out.
func (b BaseClient) POST(body interface{}, path string, out interface{}) error {
	u := *b.BaseURL
	u.Path = path

	buf := bytes.NewBuffer([]byte{})

	if body != nil {
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return err
		}
	}

	endpoint := u.String()
	b.Logf("making POST request to %s", endpoint)

	req, err := http.NewRequest(http.MethodPost, endpoint, buf)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "Application/Json")

	res, err := b.Client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	raw, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return errors.Wrap(err, "error reading response body")
	}

	b.Logf("received response: %s from POST to: %s", string(raw), endpoint)
	if reflect.TypeOf(out).Kind() == reflect.String {
		out = string(raw)
		return err
	}

	return json.Unmarshal(raw, out)
}

// StripHex removes the 0x prefix from a string.
func StripHex(s string) string {
	return hexReg.ReplaceAllString(s, "")
}
