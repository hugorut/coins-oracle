package transport

var (
	BitcoinsvAssetID = "BSV"
)

// BitcoinsvClient is the Bitcoinsv implementation of the CoinClient
type BitcoinsvClient struct {
	*BitcoinClient
}

// NewBitcoinsvClient returns a new client using os variables.
func NewBitcoinsvClient() (*BitcoinsvClient, error) {
	btcClient, err := newBTCClient(getNodeURL("BITCOINSV_URL"))
	if err != nil {
		return nil, err
	}

	return &BitcoinsvClient{
		BitcoinClient: &BitcoinClient{
			AssetID: BitcoinsvAssetID,
			Client:  btcClient,
		},
	}, nil
}
