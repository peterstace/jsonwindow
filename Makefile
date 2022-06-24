.PHONY: build
build:
	go install ./...
	go test ./...
	go vet ./...

.PHONY: stringer
stringer:
	go install golang.org/x/tools/cmd/stringer@latest
	go generate