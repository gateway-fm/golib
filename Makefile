install-upgrade-lint:
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin

lint: install-upgrade-lint
	golangci-lint run ./...

deps:
	go mod tidy

test:
	go clean --testcache && go test -count 1 -race ./...
