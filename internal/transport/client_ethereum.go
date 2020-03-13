package transport

import (
	"context"
	"github.com/hugorut/coins-oracle/pkg/transport"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
)

var (
	EthereumAssetID = "ETH"
)

// EthereumClient is the ethereum implementation of the CoinClient
type EthereumClient struct {
	AssetID string
	Client  *ethclient.Client
}

// NewEthereumClient returns a new client using the rpc endpoint given in os.
func NewEthereumClient() (*EthereumClient, error) {
	ethRpc, err := ethclient.Dial(getNodeURL("ETHEREUM_URL"))
	if err != nil {
		return nil, err
	}

	return &EthereumClient{AssetID: EthereumAssetID, Client: ethRpc}, nil
}

// GetInfo attempts to get standardised coin info from multiple rpc calls.
func (e EthereumClient) GetInfo() (*transport.CoinState, error) {
	chain, err := e.Client.NetworkID(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "error fetching network id for ethereum node")
	}

	current, err := e.Client.BlockByNumber(context.Background(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "error fetching latest block number for ethereum node")
	}

	hash := current.ReceiptHash().String()
	return &transport.CoinState{
		Data: transport.CoinData{
			Chain:        chain.String(),
			BlockHeight:  int(current.Number().Int64()),
			CurrentBlock: hash,
		},
	}, nil
}

func (e EthereumClient) GetBalance(addr string) (*transport.Balance, error) {
	am, err := e.Client.BalanceAt(context.Background(), common.HexToAddress(addr), nil)
	if err != nil {
		return nil, errors.Wrapf(err, "error getting balance for addr: %s", addr)
	}

	assetID := e.AssetID
	if assetID == "" {
		assetID = EthereumAssetID
	}

	return &transport.Balance{
		Data: transport.BalanceData{
			Assets: []transport.Asset{
				{
					Asset:   assetID,
					Balance: am.String(),
				},
			},
		},
	}, nil
}

func (e EthereumClient) GetTransactionByHash(hash string) (*transport.TransactionResp, error) {
	block, err := e.Client.BlockByNumber(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	tx, _, err := e.Client.TransactionByHash(context.Background(), common.HexToHash(hash))
	if err != nil {
		return nil, errors.Wrapf(err, "error getting transaction for hash: %s", hash)
	}

	chainId, _ := e.Client.ChainID(context.Background())
	msg, err := tx.AsMessage(types.NewEIP155Signer(chainId))
	if err != nil {
		return nil, errors.Wrap(err, "error converting eth transaction to message")
	}

	r, err := e.Client.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		return nil, err
	}

	confirmed := block.Number().Int64() - r.BlockNumber.Int64()

	return &transport.TransactionResp{
		Data: struct {
			Transaction transport.Transaction `json:"transaction"`
		}{
			Transaction: transport.Transaction{
				ID:    hash,
				From:  msg.From().String(),
				To:    tx.To().String(),
				Value: tx.Value().String(),
				Confirmations: transport.Confirmations{
					Threshold: transport.ConfirmThresholdValue,
					Confirmed: confirmed >= *transport.ConfirmThresholdValue,
					Value:     &confirmed,
				},
			},
		},
	}, nil
}
