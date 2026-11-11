GO ?= go
TINYGO ?= tinygo
WORKER_GO_TOOLCHAIN ?= go1.24.13

.PHONY: generate zenn-check worker-generate worker-release-build run test build worker-build

generate:
	$(GO) run github.com/a-h/templ/cmd/templ@v0.3.1001 generate

zenn-check:
	$(GO) run ./cmd/zenncheck

worker-generate: generate zenn-check
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
# The Cloudflare build command installs TinyGo 0.38.0, which supports Go up to 1.24.
worker-build: export GOTOOLCHAIN = $(WORKER_GO_TOOLCHAIN)
worker-build:
	$(GO) run github.com/syumai/workers/cmd/workers-assets-gen
	worker_goroot=$$($(GO) env GOROOT) && PATH="$$worker_goroot/bin:$$PATH" GOROOT="$$worker_goroot" $(TINYGO) build -tags tinygo -o ./build/app.wasm -target wasm -no-debug -opt=z ./cmd/diary
