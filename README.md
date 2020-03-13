# Coins Oracle

A lambda function which normalises many crypto currency APIs into a common interface. This allows you to query transactions and wallets using a singular API.

## Development

Running Coins Oracle Locally requires:
 
 * [go 1.13+](https://golang.org/doc/install#install)
 * [aws sam](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html)
 
once these are installed you can run `go get ./...` to install all of the required packages.

## Adding A New Coin Client
In order to simplify local development `coins-oracle` comes with a new client generator. To use is run the following command with your given coin name and asset id e.g. Bitcoin & BTC.

```
$ go run ./cmd/coingen/main.go -name=<coin_name> -asset_id=<coin_symbol>
```

After running this command the `coingen` binary will generate corresponding `client_<coin_name>.go` and `client_<coin_name>_test.go` files under `internal/transport`. The client client file comes with some boilerplate code added and the test file has some pre-written assertions that you'll need to configure and implement to get passing. 

The command also generates some fixture files for the tests under `internal/test/fixture/<coin_name>`, you should fill these in with the expected json responses from your mock coin server for your tests.

## Testing

Tests are written with `ginkgo` and use `gomega` for assertions. All test files must be placed under the `_test` namespace and whitebox testing is to be avoided.

Each coin has test file associated with it. These tests are integration tests and use a mock server to generate JSON responses for each coin. This functionality is provided by the `pkg/test` package. The `Server` allows you to easily mock expected calls like such:

```go
package client_test

import (
    "net/http"
    "testing"

    "github.com/hugorut/coins-oracle/pkg/test"
)

func TestClientMethod(t *testing.T) {
    mockServer := test.NewTestServer(t)

    
    client := SomeClient{URL: mockServer.HttpTest.URL}

    mockServer.Expect(test.ExpectedCall{
        Path:   "/",
        Method: http.MethodPost,
        Headers: map[string]string{
            "Content-Type":  "Application/Json",
            "Authorization": "Basic " + test.BasicAuth("user", "pass"),
        },
        Body:         `{"test": "request", "params": "go here"}`,
        Response:     `{"this is": "the response", "you want to return"}`,
        ResponseCode: http.StatusOK,
    }).Then(...).Then(...) // you can chain as many expectations as you want

    client.Do()

    // the test will fail if any of the server
    // expectations fail, and give a verbose output
}

```

## Running Lambda locally

The project comes with a `SAM` template file included at `sam.yaml` this is purely used for local testing purposes as deploying to prod is defined through terraform. The project makefile provides a handy `make run-gateway` cmd which starts a local gateway forwarding requests to a newly built lambda function.

You can test that your local setup has been a success by curling the `/ping` route on cmd start.

```
$ curl 127.0.0.1:3000/ping
```

will give an output:
 
```json
{"message":"pong"}
```
