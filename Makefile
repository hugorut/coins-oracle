clean:
	- rm main
	- rm main.zip

build:
	- GOOS=linux go build -o main ./cmd/coins-oracle

zip: clean build
	- zip main.zip main

run-gateway: clean build zip
	- sam local start-api -t sam.yaml

test:
	- ginkgo test ./...