package transport

var (
	BitcoinCashAssetID = "BCH"
)

// BitcoinCashClient is the Bitcoincash implementation of the CoinClient
type BitcoinCashClient struct {
	*BitcoinClient
}

// NewBitcoincashClient returns a new client using os variables.
func NewBitcoincashClient() (*BitcoinCashClient, error) {
	btcClient, err := newBTCClient(getNodeURL("BITCOINCASH_URL"))
	if err != nil {
		return nil, err
	}

	return &BitcoinCashClient{
		BitcoinClient: &BitcoinClient{
			AssetID: BitcoinCashAssetID,
			Client:  btcClient,
		},
	}, nil
}
