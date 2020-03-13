package transport

var (
	LitecoinAssetID = "LTC"
)

// LitecoinClient is the Litecoin implementation of the CoinClient
type LitecoinClient struct {
	*BitcoinClient
}

// NewLitecoinClient returns a new client using os variables.
func NewLitecoinClient() (*LitecoinClient, error) {
	btcClient, err := newBTCClient(getNodeURL("LITECOIN_URL"))
	if err != nil {
		return nil, err
	}

	return &LitecoinClient{
		BitcoinClient: &BitcoinClient{
			AssetID: LitecoinAssetID,
			Client:  btcClient,
		},
	}, nil
}
