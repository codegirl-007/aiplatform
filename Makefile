APP_NAME := aiplatform

.PHONY: help dev build clean test coverage lint fmt vet tigerlint

help:
	@printf "Targets:\n"
	@printf "  dev         Run Wails dev mode\n"
	@printf "  build       Build Wails production app\n"
	@printf "  clean       Remove build artifacts\n"
	@printf "  test        Run Go tests\n"
	@printf "  coverage    Run tests and show coverage in terminal\n"
	@printf "  lint        Run go vet and custom tigerlint\n"
	@printf "  fmt         Run gofmt on Go files\n"
	@printf "  vet         Run go vet\n"
	@printf "  tigerlint   Run Tiger Beetle linter\n"

dev:
	wails dev

build:
	wails build

clean:
	rm -rf build

test:
	go test ./...

coverage:
	go test ./... -coverpkg=./... -coverprofile=/tmp/coverage.out -covermode=atomic && \
	printf "\n=== Coverage Summary ===\n" && \
	go tool cover -func=/tmp/coverage.out | tail -1 && \
	printf "\n=== Detailed Coverage ===\n" && \
	go tool cover -func=/tmp/coverage.out && \
	rm -f /tmp/coverage.out

lint: vet tigerlint

fmt:
	gofmt -w $$(go list -f '{{.Dir}}' ./... | xargs -I {} find {} -name '*.go')

vet:
	go vet ./...

tigerlint:
	go run ./cmd/tigerlint ./...
