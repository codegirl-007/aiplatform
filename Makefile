APP_NAME := aiplatform

.PHONY: help dev build clean test lint fmt vet tigerlint

help:
	@printf "Targets:\n"
	@printf "  dev         Run Wails dev mode\n"
	@printf "  build       Build Wails production app\n"
	@printf "  clean       Remove build artifacts\n"
	@printf "  test        Run Go tests\n"
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

lint: vet tigerlint

fmt:
	gofmt -w $$(go list -f '{{.Dir}}' ./... | xargs -I {} find {} -name '*.go')

vet:
	go vet ./...

tigerlint:
	go run ./cmd/tigerlint ./...
