.PHONY: lint
lint:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.4.0
	@go mod download
	go mod tidy -diff > /dev/null
	golangci-lint run ./... --timeout 5m

.PHONY: test
test:
	go test ./... -gcflags 'all=-l' -failfast -timeout 20s -coverprofile .test-coverage.txt

.PHONY: bench
bench:
	go test -bench=. ./... -timeout 20s -run=^$$

.PHONY: coverage-report
coverage-report: test
	go tool cover -func=.test-coverage.txt
