package transport

var (
	DogecoinAssetID = "DOGE"
)

// DogecoinClient is the Dogecoin implementation of the CoinClient
type DogecoinClient struct {
	*BitcoinClient
}

// NewDogecoinClient returns a new client using os variables.
func NewDogecoinClient() (*DogecoinClient, error) {
	btcClient, err := newBTCClient(getNodeURL("DOGECOIN_URL"))
	if err != nil {
		return nil, err
	}

	return &DogecoinClient{
		BitcoinClient: &BitcoinClient{
			AssetID: DogecoinAssetID,
			Client:  btcClient,
		},
	}, nil
}
