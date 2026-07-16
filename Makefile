GO ?= go

.PHONY: generate worker-generate run test build worker-build

generate:
	$(GO) tool templ generate

worker-generate: generate
	$(GO) run ./cmd/workergen

run: generate
	$(GO) run ./cmd/diary

test: generate
	$(GO) test ./...

build: generate
	$(GO) build ./...

worker-build: worker-generate
	$(GO) run github.com/syumai/workers/cmd/workers-assets-gen
	PATH=$$($(GO) env GOROOT)/bin:$$PATH GOROOT=$$($(GO) env GOROOT) tinygo build -tags tinygo -o ./build/app.wasm -target wasm -no-debug ./cmd/diary
