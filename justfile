check: fmt build

fmt:
	go fmt ./...

build:
	go build ./...

run:
	go run cmd/kdl/main.go | less
