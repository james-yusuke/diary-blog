GO ?= go
TINYGO ?= tinygo

.PHONY: generate worker-generate worker-release-build run test build worker-build

generate:
	$(GO) run github.com/a-h/templ/cmd/templ@v0.3.1001 generate

worker-generate: generate
	$(GO) run ./cmd/workergen

# Regenerate the checked-in Worker content snapshot, then build it. Use this
# locally after changing Markdown; Cloudflare builds use worker-build so they
# never need credentials for the private Zenn repository.
worker-release-build: worker-generate worker-build

run: generate
	$(GO) run ./cmd/diary

test: generate
	$(GO) test ./...

build: generate
	$(GO) build ./...

# Build the committed content snapshot. Do not make this depend on
# worker-generate: Cloudflare receives the generated snapshot in Git and must
# not overwrite it when the private Zenn checkout is unavailable.
worker-build:
	$(GO) run github.com/syumai/workers/cmd/workers-assets-gen
	PATH=$$($(GO) env GOROOT)/bin:$$PATH GOROOT=$$($(GO) env GOROOT) $(TINYGO) build -tags tinygo -o ./build/app.wasm -target wasm -no-debug -opt=z ./cmd/diary
