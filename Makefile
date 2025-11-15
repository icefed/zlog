COVERAGE_DIR?=coverage

.PHONY: all
all: test

fmt:
	go fmt ./...

lint:
	golangci-lint run

test: fmt
	go test -v ./...

test-cover: fmt
	mkdir -p $(COVERAGE_DIR)
	go test -v -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	go tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html

bench:
	@cd benchmarks && go test -bench=. -benchmem | tee bench.txt

clean:
	rm -rf $(COVERAGE_DIR)
