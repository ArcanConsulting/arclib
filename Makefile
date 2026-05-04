.PHONY: test-go test-js test-kt test-cross test-all lint security coverage fmt check ci clean

GO := go
GOTEST := $(GO) test
GOVET := $(GO) vet
COVERAGE_MIN := 95

## Testing

test-go:
	$(GOTEST) -race -v ./...

test-js:
	cd js && npm test

test-kt:
	cd kt && ./gradlew test

test-cross:
	$(GOTEST) -race -v -tags=crossvalidation ./testdata/...

test-all: test-go test-js test-kt test-cross

## Quality

lint:
	golangci-lint run ./...

security:
	gosec ./...
	govulncheck ./...

fmt:
	gofmt -w .
	goimports -w .

## Coverage

coverage:
	$(GOTEST) -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GO) tool cover -func=coverage.out
	$(GO) tool cover -html=coverage.out -o coverage.html

coverage-check: coverage
	@COVERAGE=$$($(GO) tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	echo "Coverage: $${COVERAGE}%"; \
	if [ $$(echo "$${COVERAGE} < $(COVERAGE_MIN)" | bc -l) -eq 1 ]; then \
		echo "FAIL: Coverage $${COVERAGE}% is below minimum $(COVERAGE_MIN)%"; \
		exit 1; \
	fi

## Combined

check: lint security test-go

ci: lint security coverage-check

## Cleanup

clean:
	rm -f coverage.out coverage.html
	$(GO) clean -testcache
