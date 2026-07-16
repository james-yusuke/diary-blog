.PHONY: generate worker-generate run test build worker-build

generate:
	go tool templ generate

worker-generate: generate
	go run ./cmd/workergen

run: generate
	go run ./cmd/diary

test: generate
	go test ./...

build: generate
	go build ./...

worker-build: worker-generate
	go run github.com/syumai/workers/cmd/workers-assets-gen
	tinygo build -tags tinygo -o ./build/app.wasm -target wasm -no-debug ./cmd/diary
