package transport

import (
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	EthereumclassicAssetID = "ETC"
)

// EthereumClassicClient is the EthereumClassic implementation of the CoinClient
type EthereumClassicClient struct {
	*EthereumClient
}

// NewEthereumClassicClient returns a new client using os variables.
func NewEthereumClassicClient() (*EthereumClassicClient, error) {
	ethRpc, err := ethclient.Dial(getNodeURL("ETHEREUMCLASSIC_URL"))
	if err != nil {
		return nil, err
	}

	return &EthereumClassicClient{
		EthereumClient: &EthereumClient{
			AssetID: EthereumclassicAssetID,
			Client:  ethRpc,
		},
	}, nil
}
