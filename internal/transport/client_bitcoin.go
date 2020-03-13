package transport

import (
	"bytes"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btclog"
	"github.com/btcsuite/btcutil"
	"github.com/pkg/errors"

	"github.com/hugorut/coins-oracle/pkg/transport"
)

var (
	BitcoinAssetID = "BTC"

	ErrorAlreadyImported = errors.New("address already imported")
)

// btcStdLogger implements the BTCLogger interface and directs output to stdout.
type btcStdLogger struct {
	log   *log.Logger
	level btclog.Level
}

func (b btcStdLogger) printF(meth, format string, params ...interface{}) {
	s := fmt.Sprintf(format, params...)
	b.log.Printf("[BTC CLIENT] %s ==> %s", meth, s)
}

func (b btcStdLogger) Tracef(format string, params ...interface{}) {
	b.printF("TRACE", format, params...)
}
func (b btcStdLogger) Debugf(format string, params ...interface{}) {
	b.printF("DEBUG", format, params...)
}
func (b btcStdLogger) Infof(format string, params ...interface{}) { b.printF("INFO", format, params...) }
func (b btcStdLogger) Warnf(format string, params ...interface{}) { b.printF("WARN", format, params...) }
func (b btcStdLogger) Errorf(format string, params ...interface{}) {
	b.printF("ERROR", format, params...)
}
func (b btcStdLogger) Criticalf(format string, params ...interface{}) {
	b.printF("CRITICAL", format, params...)
}

func (b btcStdLogger) Trace(v ...interface{})    { b.log.Println("TRACE"); b.log.Println(v...) }
func (b btcStdLogger) Debug(v ...interface{})    { b.log.Println("DEBUG"); b.log.Println(v...) }
func (b btcStdLogger) Info(v ...interface{})     { b.log.Println("INFO"); b.log.Println(v...) }
func (b btcStdLogger) Warn(v ...interface{})     { b.log.Println("WARN"); b.log.Println(v...) }
func (b btcStdLogger) Error(v ...interface{})    { b.log.Println("ERROR"); b.log.Println(v...) }
func (b btcStdLogger) Critical(v ...interface{}) { b.log.Println("CRITICAL"); b.log.Println(v...) }

func (b btcStdLogger) Level() btclog.Level          { return b.level }
func (b *btcStdLogger) SetLevel(level btclog.Level) { b.level = level }

// btcStrAddr conforms to the bctutil.Address interface
// but simply returns the base addr string.
type btcStrAddr struct{ addr string }

func (b btcStrAddr) String() string                 { return b.addr }
func (b btcStrAddr) EncodeAddress() string          { return b.addr }
func (b btcStrAddr) ScriptAddress() []byte          { return []byte(b.addr) }
func (b btcStrAddr) IsForNet(*chaincfg.Params) bool { return true }

// BitcoinClient is the Bitcoin implementation of the CoinClient
type BitcoinClient struct {
	AssetID string
	Client  *rpcclient.Client
}

// NewBitcoinClient returns a new client using os variables.
func NewBitcoinClient() (*BitcoinClient, error) {
	btcClient, err := newBTCClient(getNodeURL("BITCOIN_URL"))
	if err != nil {
		return nil, err
	}

	return &BitcoinClient{
		AssetID: BitcoinAssetID,
		Client:  btcClient,
	}, nil
}

func newBTCClient(addr string) (*rpcclient.Client, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	connCfg := &rpcclient.ConnConfig{
		Host:         u.Host,
		User:         os.Getenv("RPC_USER"),
		Pass:         os.Getenv("RPC_PASS"),
		HTTPPostMode: true,
		DisableTLS:   true,
	}

	rpcclient.UseLogger(&btcStdLogger{log: log.New(os.Stderr, "", log.LstdFlags)})
	return rpcclient.New(connCfg, nil)
}

// GetInfo attempts to get standardised coin info from multiple rpc calls.
func (b BitcoinClient) GetInfo() (*transport.CoinState, error) {
	res, err := b.Client.GetBlockChainInfo()
	if err != nil {
		return nil, err
	}

	return &transport.CoinState{
		Data: transport.CoinData{
			Chain:        res.Chain,
			BlockHeight:  int(res.Blocks),
			CurrentBlock: res.BestBlockHash,
		},
	}, nil
}

// GetBalance returns the balance of the address.
func (b BitcoinClient) GetBalance(addr string) (*transport.Balance, error) {
	unspent, err := b.Client.ListUnspentMinMaxAddresses(1, 9999999, []btcutil.Address{btcStrAddr{addr: addr}})
	if err != nil {
		return nil, errors.Wrap(err, "error listing unspent for given addr")
	}

	var total float64
	for _, value := range unspent {
		total += value.Amount
	}

	return &transport.Balance{
		Data: transport.BalanceData{
			Assets: []transport.Asset{
				{
					Asset:   b.AssetID,
					Balance: fmt.Sprintf("%f", total),
				},
			},
		},
	}, nil
}

// GetTransactionByHash returns the transaction stored at the given hash.
func (b BitcoinClient) GetTransactionByHash(hash string) (*transport.TransactionResp, error) {
	raw, err := b.getTransaction(hash)
	if err != nil {
		return nil, errors.Wrapf(err, "error getting transaction from initial input hash: %s", hash)
	}

	sendingTx := raw.Vin[0]

	from, err := b.getTransaction(sendingTx.Txid)
	if err != nil {
		return nil, errors.Wrapf(err, "error getting transaction for input transaction: %s", sendingTx.Txid)
	}

	var value float64
	for _, v := range raw.Vout {
		value += v.Value
	}

	confirmationsValue := transport.NewInt64(int64(raw.Confirmations))

	return &transport.TransactionResp{
		Data: struct {
			Transaction transport.Transaction `json:"transaction"`
		}{
			Transaction: transport.Transaction{
				From:  from.Vout[sendingTx.Vout].ScriptPubKey.Addresses[0],
				To:    raw.Vout[0].ScriptPubKey.Addresses[0],
				ID:    raw.Hash,
				Value: fmt.Sprintf("%f", value),
				Confirmations: transport.Confirmations{
					Threshold: transport.ConfirmThresholdValue,
					Confirmed: *confirmationsValue >= *transport.ConfirmThresholdValue,
					Value:     confirmationsValue,
				},
			},
		},
	}, nil
}

func (b BitcoinClient) getTransaction(hash string) (*btcjson.TxRawResult, error) {
	chainH, err := chainhash.NewHashFromStr(hash)
	if err != nil {
		return nil, errors.Wrap(err, "error generating a chain hash from given hash")
	}

	msg, err := b.Client.GetRawTransaction(chainH)
	if err != nil {
		return nil, errors.Wrap(err, "error getting raw transaction")
	}

	buf := bytes.NewBuffer([]byte{})
	err = msg.MsgTx().Serialize(buf)
	if err != nil {
		return nil, errors.Wrap(err, "could not serialize transaction msg")
	}

	raw, err := b.Client.DecodeRawTransaction(buf.Bytes())
	if err != nil {
		return nil, errors.Wrap(err, "error decoding transaction")
	}

	return raw, nil
}

// ImportAddress imports the given address. This will reindex the chain, which may block connections, so use wisely.
func (b BitcoinClient) ImportAddress(addr string) error {
	err := b.importAddress(addr)
	if err == ErrorAlreadyImported || err == nil {
		return nil
	}

	err = b.importAddress(addr)
	if err == ErrorAlreadyImported || err == nil {
		return nil
	}

	return err
}

// importAddress wraps the btcClients method in with a timeout since this method.
// This is done as the internals don't have a default timeout and this importAddress
// normally blocks for a considerable with the first call. It only returns a response
// after successful import, which can take hours.
//
// On second call it will return a address already imported error. So the solution is to
// cancel the first request and try again to make sure the import is kicked off.
func (b BitcoinClient) importAddress(addr string) error {
	errChn := make(chan error)
	go func() {
		errChn <- b.Client.ImportAddress(addr)
		close(errChn)
	}()

	for {
		select {
		case err := <-errChn:
			btcErr, ok := err.(*btcjson.RPCError)
			if !ok {
				return err
			}

			if btcErr.Code == btcjson.ErrRPCWallet {
				return ErrorAlreadyImported
			}

			return err
		case <-time.After(transport.DefaultClientTimeout):
			return errors.New("timeout")
		}
	}
}
