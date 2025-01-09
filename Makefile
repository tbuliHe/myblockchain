build:
	go build -o ./bin/myblockchain

run: build
	./bin/myblockchain

test:
	go test -v ./...