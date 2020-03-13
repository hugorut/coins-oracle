package transport

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"

	"github.com/hugorut/coins-oracle/pkg/transport"
)

var (
	TetherAssetID    = "USDT"
	OxAssetID        = "ZRX"
	BATAssetID       = "BAT"
	ChainLinkAssetID = "LINK"
	IconAssetID      = "ICX"
	MakerAssetID     = "MKR"
	OmiseGoAssetID   = "OMG"
	VeChainAssetID   = "VEN"
	ZilliqaAssetID   = "ZIL"

	// ERC20Tokens are a map of erc20 token config.
	ERC20Tokens = map[string]ERC20Config{
		TetherAssetID: {
			AssetID:      TetherAssetID,
			ContractAddr: "0xdAC17F958D2ee523a2206206994597C13D831ec7",
		},
		OxAssetID: {
			AssetID:      OxAssetID,
			ContractAddr: "0xE41d2489571d322189246DaFA5ebDe1F4699F498",
		},
		BATAssetID: {
			AssetID:      BATAssetID,
			ContractAddr: "0x0D8775F648430679A709E98d2b0Cb6250d2887EF",
		},
		ChainLinkAssetID: {
			AssetID:      ChainLinkAssetID,
			ContractAddr: "0x514910771AF9Ca656af840dff83E8264EcF986CA",
		},
		IconAssetID: {
			AssetID:      IconAssetID,
			ContractAddr: "0xb5a5f22694352c15b00323844ad545abb2b11028",
		},
		MakerAssetID: {
			AssetID:      MakerAssetID,
			ContractAddr: "0x9f8f72aa9304c8b593d555f12ef6589cc3a579a2",
		},
		OmiseGoAssetID: {
			AssetID:      OmiseGoAssetID,
			ContractAddr: "0xd26114cd6EE289AccF82350c8d8487fedB8A0C07",
		},
		VeChainAssetID: {
			AssetID:      VeChainAssetID,
			ContractAddr: "0xd850942ef8811f2a866692a623011bde52a462c1",
		},
		ZilliqaAssetID: {
			AssetID:      ZilliqaAssetID,
			ContractAddr: "0x05f4a42e251f2d52b8ed15E9FEdAacFcEF1FAD27",
		},
	}

	// this is the same as writing balanceOf(address) for the contract
	balanceOfEncStr = "0x70a08231"
	etherScan, _    = url.Parse("https://api.etherscan.io")
)

// ERC20Config defines a struct to hold run parameters for an erc 20 coin.
type ERC20Config struct {
	AssetID      string
	ContractAddr string
}

// ERC20ContractTxData holds information about the contract token transfer
// this is decoded from a transaction input.
type ERC20ContractTxData struct {
	Token  common.Address
	To     common.Address
	Value  *big.Int
	Amount *big.Int
}

// ERC20Client is the ERC20 implementation of the CoinClient
type ERC20Client struct {
	AssetID      string
	ContractAddr *common.Address
	EthClient    *ethclient.Client
	ABIClient    transport.BaseClient

	ABIMap map[string]abi.ABI
	MU     *sync.Mutex
}

// ABIResult maps the JSON result from an ABI lookup on etherscan.
type ABIResult struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  string `json:"result"`
}

// NewERC20Client returns a new client using os variables.
func NewERC20Client(token string) (*ERC20Client, error) {
	config, ok := ERC20Tokens[token]
	if !ok {
		return nil, fmt.Errorf("ERC20 token: %s not found, please add to config map", token)
	}

	ethRpc, err := ethclient.Dial(getNodeURL("ETHEREUM_URL"))
	if err != nil {
		return nil, errors.Wrap(err, "error initializing base ethereum client for erc20 client")
	}

	addr := common.HexToAddress(config.ContractAddr)

	return &ERC20Client{
		AssetID:      config.AssetID,
		ContractAddr: &addr,
		EthClient:    ethRpc,
		ABIClient: transport.BaseClient{
			BaseURL: etherScan,
			Client:  &http.Client{Timeout: time.Second * time.Duration(10)},
			Log:     log.New(os.Stdout, "", log.LstdFlags),
		},
		ABIMap: map[string]abi.ABI{},
		MU:     &sync.Mutex{},
	}, nil
}

// GetInfo attempts to get standardised coin info from multiple rpc calls.
func (e ERC20Client) GetInfo() (*transport.CoinState, error) {
	return &transport.CoinState{
		Data: transport.CoinData{
			Chain:        "main",
			CurrentBlock: "ERC20",
		},
	}, nil
}

// GetBalance returns the balance of the address.
func (e ERC20Client) GetBalance(addr string) (*transport.Balance, error) {
	data, _ := hexutil.Decode(balanceOfEncStr + "000000000000000000000000" + transport.StripHex(addr))
	msg := ethereum.CallMsg{
		To:   e.ContractAddr,
		Data: data,
	}

	b, err := e.EthClient.CallContract(context.Background(), msg, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error fetching contract balance for contract")
	}

	if len(b) == 0 {
		return nil, errors.New("error fetching contract balance, returned nil bytes slice")
	}

	buf := bytes.NewBuffer([]byte("0x"))
	hexString := hexutil.Bytes(b).String()

	for _, value := range transport.StripHex(hexString) {
		if value != 48 {
			buf.WriteRune(value)
		}
	}

	res, err := hexutil.DecodeBig(buf.String())
	if err != nil {
		return nil, errors.Wrapf(err, "error decoding balance string: %s, original hex: %s", buf.String(), hexString)
	}

	return &transport.Balance{
		Data: transport.BalanceData{
			Assets: []transport.Asset{
				{
					Asset:   e.AssetID,
					Balance: res.String(),
				},
			},
		},
	}, nil
}

// GetTransactionByHash returns the transaction stored at the given hash.
func (e *ERC20Client) GetTransactionByHash(hash string) (*transport.TransactionResp, error) {
	block, err := e.EthClient.BlockByNumber(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	tx, _, err := e.EthClient.TransactionByHash(context.Background(), common.HexToHash(hash))
	if err != nil {
		return nil, errors.Wrapf(err, "error getting transaction for hash: %s", hash)
	}

	chainId, _ := e.EthClient.ChainID(context.Background())
	msg, err := tx.AsMessage(types.NewEIP155Signer(chainId))
	if err != nil {
		return nil, errors.Wrap(err, "error converting eth transaction to message")
	}

	r, err := e.EthClient.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		return nil, err
	}

	txABI, err := e.findABI(tx.To())
	if err != nil {
		return nil, err
	}

	confirmed := block.Number().Int64() - r.BlockNumber.Int64()

	txInput := hexutil.Bytes(tx.Data()).String()
	decodedSig, _ := hex.DecodeString(txInput[2:10])

	method, err := txABI.MethodById(decodedSig)
	if err != nil {
		return nil, errors.Wrap(err, "error getting abi method by id")
	}

	decodedData, _ := hex.DecodeString(txInput[10:])

	var data ERC20ContractTxData
	err = method.Inputs.Unpack(&data, decodedData)
	if err != nil {
		return nil, errors.Wrap(err, "error unpacking transaction data")
	}

	to := data.To.String()
	if len(data.To.Bytes()) == 0 {
		to = data.Token.String()
	}

	var value string
	if data.Value == nil {
		value = data.Amount.String()
	} else {
		value = data.Value.String()
	}

	return &transport.TransactionResp{
		Data: struct {
			Transaction transport.Transaction `json:"transaction"`
		}{
			Transaction: transport.Transaction{
				ID:    hash,
				From:  msg.From().String(),
				To:    to,
				Value: value,
				Confirmations: transport.Confirmations{
					Threshold: transport.ConfirmThresholdValue,
					Confirmed: confirmed >= *transport.ConfirmThresholdValue,
					Value:     &confirmed,
				},
			},
		},
	}, nil
}

func (e *ERC20Client) findABI(contract *common.Address) (abi.ABI, error) {
	e.MU.Lock()
	defer e.MU.Unlock()

	addr := contract.String()

	if txABI, ok := e.ABIMap[addr]; ok {
		return txABI, nil
	}

	var abiRes ABIResult
	err := e.ABIClient.GET("/api", map[string]string{
		"module":  "contract",
		"action":  "getabi",
		"address": addr,
	}, &abiRes)
	if err != nil {
		return abi.ABI{}, errors.Wrap(err, "error getting abi for contract")
	}

	txABI, err := abi.JSON(strings.NewReader(abiRes.Result))
	if err != nil {
		return abi.ABI{}, errors.Wrap(err, "error parsing ABI json")
	}

	e.ABIMap[addr] = txABI
	return txABI, nil
}
