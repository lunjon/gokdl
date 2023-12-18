check: fmt test lint

build:
	go build ./...

fmt:
	go fmt ./...

test pattern=".*":
	go test ./... -run={{ pattern }}

lint:
	go run honnef.co/go/tools/cmd/staticcheck@latest ./...
