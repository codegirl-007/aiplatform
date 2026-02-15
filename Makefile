APP_NAME := aiplatform

.PHONY: help dev build clean test coverage lint fmt vet tigerlint

help:
	@printf "Targets:\n"
	@printf "  dev         Run Wails dev mode\n"
	@printf "  build       Build Wails production app\n"
	@printf "  clean       Remove build artifacts\n"
	@printf "  test        Run Go tests\n"
	@printf "  coverage    Generate test coverage report\n"
	@printf "  lint        Run go vet and custom tigerlint\n"
	@printf "  fmt         Run gofmt on Go files\n"
	@printf "  vet         Run go vet\n"
	@printf "  tigerlint   Run Tiger Beetle linter\n"

dev:
	wails dev

build:
	wails build

clean:
	rm -rf build coverage.out coverage.html

test:
	go test ./...

coverage:
	go test ./... -coverpkg=./... -coverprofile=coverage.out
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@printf "\nWrote coverage.out and coverage.html\n"

lint: vet tigerlint

fmt:
	gofmt -w $$(go list -f '{{.Dir}}' ./... | xargs -I {} find {} -name '*.go')

vet:
	go vet ./...

tigerlint:
	go run ./cmd/tigerlint ./...
