.PHONY: build test vet fmt run

build:
	go build -o oss-radar .

test:
	go test ./...

vet:
	go vet ./...

fmt:
	test -z "$$(gofmt -l .)"

run:
	GITHUB_TOKEN=$${GITHUB_TOKEN:-} go run . -json
