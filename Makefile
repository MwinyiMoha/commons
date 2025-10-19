test:
	go test -v -cover ./...

install:
	go mod download

.PHONY: test install