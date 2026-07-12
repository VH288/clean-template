.PHONY: proto run run-reference build

proto:
	PATH="$$(go env GOPATH)/bin:$$PATH" protoc \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/sample/v1/sample.proto \
		proto/external/v1/reference.proto

run:
	go run ./cmd/server

run-reference:
	go run ./cmd/reference-server

build:
	go build -o bin/server ./cmd/server
	go build -o bin/reference-server ./cmd/reference-server
