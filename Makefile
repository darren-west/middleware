all: dependencies generate test

dependencies:
	GO111MODULE=on go mod download

test:
	GO111MODULE=on go test ./... --cover -v -tags=$(TEST_TAGS)

generate:
	GO111MODULE=on go generate ./...
