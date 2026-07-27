GO ?= go
TINYGO ?= tinygo

.PHONY: generate worker-generate run test build worker-build

generate:
	$(GO) run github.com/a-h/templ/cmd/templ@v0.3.1001 generate

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
	PATH=$$($(GO) env GOROOT)/bin:$$PATH GOROOT=$$($(GO) env GOROOT) $(TINYGO) build -tags tinygo -o ./build/app.wasm -target wasm -no-debug -opt=z ./cmd/diary
