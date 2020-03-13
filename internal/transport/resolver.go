package transport

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/hugorut/coins-oracle/pkg/transport"

	"github.com/labstack/echo"

	"github.com/stellar/go/support/errors"
)

var (
	httpReg = regexp.MustCompile("^https?://")
)

// NewResolver returns a new instance of the CoinResolver with clients initiated
func NewResolver(l echo.Logger) *CoinResolver {
	r := &CoinResolver{
		C:      make(map[string]transport.CoinClient),
		Mu:     &sync.Mutex{},
		Logger: l,
	}

	r.Register(EthereumAssetID, must(NewEthereumClient()))
	r.Register(BitcoinAssetID, must(NewBitcoinClient()))
	r.Register(EosAssetID, must(NewEosClient()))
	r.Register(RippleAssetID, must(NewRippleClient()))
	r.Register(CardanoAssetID, must(NewCardanoClient()))
	r.Register(TronAssetID, must(NewTronClient()))
	r.Register(NemAssetID, must(NewNemClient()))
	r.Register(NanoAssetID, must(NewNanoClient()))
	r.Register(NeoAssetID, must(NewNeoClient()))
	r.Register(StellarAssetID, must(NewStellarClient()))
	r.Register(BitcoinsvAssetID, must(NewBitcoinsvClient()))
	r.Register(LitecoinAssetID, must(NewLitecoinClient()))
	r.Register(BitcoinCashAssetID, must(NewBitcoincashClient()))
	r.Register(DogecoinAssetID, must(NewDogecoinClient()))
	r.Register(EthereumclassicAssetID, must(NewEthereumClassicClient()))
	r.Register(BitcoinGoldAssetID, must(NewBitcoinGoldClient()))
	r.Register(TezosAssetID, must(NewTezosClient()))
	r.Register(OntologyAssetID, must(NewOntologyClient()))
	r.Register(LiskAssetID, must(NewLiskClient()))
	r.Register(WavesAssetID, must(NewWavesClient()))
	r.Register(TetherAssetID, must(NewERC20Client(TetherAssetID)))
	r.Register(OxAssetID, must(NewERC20Client(OxAssetID)))
	r.Register(BATAssetID, must(NewERC20Client(BATAssetID)))
	r.Register(ChainLinkAssetID, must(NewERC20Client(ChainLinkAssetID)))
	r.Register(IconAssetID, must(NewERC20Client(IconAssetID)))
	r.Register(MakerAssetID, must(NewERC20Client(MakerAssetID)))
	r.Register(OmiseGoAssetID, must(NewERC20Client(OmiseGoAssetID)))
	r.Register(VeChainAssetID, must(NewERC20Client(VeChainAssetID)))
	r.Register(ZilliqaAssetID, must(NewERC20Client(ZilliqaAssetID)))
	r.Register(QtumAssetID, must(NewQtumClient()))
	r.Register(TezosAssetID, must(NewTezosClient()))
	r.Register(IotaAssetID, must(NewIotaClient()))
	r.Register(DecredAssetID, must(NewDecredClient()))

	return r
}

func must(c transport.CoinClient, err error) transport.CoinClient {
	if err != nil {
		s := reflect.TypeOf(c).String()
		log.Fatal(errors.Wrapf(err, "error creating %s", s))
	}

	return c
}

func getNodeURL(s string) string {
	v := nodeURLFromEnv(s)

	log.Printf("executing os value %s with: %s\n", s, v)

	return v
}

func nodeURLFromEnv(s string) string {
	v := os.Getenv(s)
	if v == "" {
		return "http://localhost"
	}

	if httpReg.MatchString(v) {
		return v
	}

	return "http://" + v
}

// CoinNode represents a running blockchain coin.
type CoinNode struct {
	AssetId string              `json:"assetId"`
	Running bool                `json:"running"`
	Info    *transport.CoinData `json:"info,omitempty"`
}

// Resolver defines an interface which can register and pull different clients.
type Resolver interface {
	Register(name string, client transport.CoinClient) *CoinResolver
	Get(name string) (transport.CoinClient, error)
	GetNodes(info bool) []CoinNode
}

// CoinResolver is a lookup container for registering and calling
// different lambdas based on their coin name
type CoinResolver struct {
	C      map[string]transport.CoinClient
	Mu     *sync.Mutex
	Logger echo.Logger
}

// Register registers a new client using the given name
func (r *CoinResolver) Register(name string, client transport.CoinClient) *CoinResolver {
	r.Mu.Lock()
	defer r.Mu.Unlock()

	r.C[strings.ToLower(name)] = client
	return r
}

// Get returns a CoinClient registered at the given name.
// If a client is not registered it will return a not found error.
func (r CoinResolver) Get(name string) (transport.CoinClient, error) {
	if v, ok := r.C[strings.ToLower(name)]; ok {
		return v, nil
	}

	return nil, fmt.Errorf("could not find client named: %s, have you registered the client", name)
}

// GetNodes returns a list of registered nodes.
// If info parameter is passed CoinResolver will attempt to get up to date information about the node.
func (r CoinResolver) GetNodes(info bool) []CoinNode {
	wg := sync.WaitGroup{}

	list := make([]CoinNode, len(r.C))

	var i int
	for n, c := range r.C {
		coin := CoinNode{
			AssetId: n,
			Running: true,
		}

		// if info let's start a new go routine which is non blocking to fetch coin info.
		if info {
			r.Logger.Printf("executing a GetInfo request for CoinClient: %s\n	", n)
			wg.Add(1)

			go func(asset string, index int, client transport.CoinClient) {
				cs, err := client.GetInfo()
				r.Logger.Printf("%s client returned response, err: %s, res: %+v\n", asset, err, cs)
				if err != nil {
					wg.Done()
					return
				}

				list[index].Info = &cs.Data
				wg.Done()
			}(n, i, c)
		}

		list[i] = coin
		i++
	}

	if info {
		r.Logger.Print("waiting for coin nodes")
		wg.Wait()
	}

	return list
}
