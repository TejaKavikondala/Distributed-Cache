PROTOC_GEN_GO := $(shell go env GOPATH)/bin/protoc-gen-go
PROTOC_GEN_GO_GRPC := $(shell go env GOPATH)/bin/protoc-gen-go-grpc

all: proto build

proto:
	protoc --go_out=. --go-grpc_out=. proto/cache.proto

build:
	go build -o cache-node ./cmd/node

test:
	go test ./...

clean:
	rm -rf cache-node
