check: fmt test

fmt:
	go fmt ./...

build:
	go build ./...

test pattern=".*":
	go test ./... -run={{ pattern }}
