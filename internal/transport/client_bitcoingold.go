package transport

var (
	BitcoinGoldAssetID = "BTG"
)

// BitcoinGoldClient is the BitcoinGold implementation of the CoinClient
type BitcoinGoldClient struct {
	*BitcoinClient
}

// NewBitcoinGoldClient returns a new client using os variables.
func NewBitcoinGoldClient() (*BitcoinGoldClient, error) {
	btcClient, err := newBTCClient(getNodeURL("BITCOINGOLD_URL"))
	if err != nil {
		return nil, err
	}

	return &BitcoinGoldClient{
		BitcoinClient: &BitcoinClient{
			AssetID: BitcoinGoldAssetID,
			Client:  btcClient,
		},
	}, nil
}

