lint:
	golangci-lint run ./...

protoc:
	protoc --go_out=. ./example/proto/helloworld.proto
