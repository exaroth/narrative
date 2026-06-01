.PHONY: build
build:
	go build -o ./build/narrative ./cmd/narrative/main.go

.PHONY: build-debugger
build-debugger:
	go build -o debugger ./cmd-aux/debugger/main.go

.PHONY: run
run:
	DEBUG=1 go run ./cmd/narrative/main.go

# Setup test env
.PHONY: test-setup
test-setup:
	go install github.com/kyoh86/richgo@latest
	go install github.com/jstemmer/go-junit-report@latest

.PHONY: test
test:
	richgo test ./... -mod=readonly -v

.PHONY: test-cov
test-cov:
	richgo test -v -race -coverpkg=./... -coverprofile=coverage.txt ./... -mod=readonly

.PHONY: get-kitten-mini
get-kitten-mini:
	mkdir -p ./models/kitten
	wget "https://huggingface.co/KittenML/kitten-tts-mini-0.8/resolve/main/kitten_tts_mini_v0_8.onnx?download=true" -O ./models/kitten/kitten.onnx
	wget "https://huggingface.co/KittenML/kitten-tts-mini-0.8/resolve/main/voices.npz?download=true" -O ./models/kitten/voices.npz

.PHONY: get-kitten-micro
get-kitten-micro:
	mkdir -p ./models/kitten
	wget "https://huggingface.co/KittenML/kitten-tts-micro-0.8/resolve/main/kitten_tts_micro_v0_8.onnx?download=true" -O ./models/kitten/kitten.onnx
	wget "https://huggingface.co/KittenML/kitten-tts-micro-0.8/resolve/main/voices.npz?download=true" -O ./models/kitten/voices.npz

.PHONY: get-kitten-nano
get-kitten-nano:
	mkdir -p ./models/kitten
	wget "https://huggingface.co/KittenML/kitten-tts-nano-0.8/resolve/main/kitten_tts_nano_v0_8.onnx?download=true" -O ./models/kitten/kitten.onnx
	wget "https://huggingface.co/KittenML/kitten-tts-nano-0.8/resolve/main/voices.npz?download=true" -O ./models/kitten/voices.npz


.PHONY: get-lib-linux
get-lib-linux:
	mkdir -p ./libs/ /tmp/narrative-onnx-linux
	wget "https://github.com/microsoft/onnxruntime/releases/download/v1.25.1/onnxruntime-linux-x64-1.25.1.tgz" -O /tmp/narrative-onnx-linux/lib.tgz
	tar -xvzf /tmp/narrative-onnx-linux/lib.tgz  -C /tmp/narrative-onnx-linux --strip-components=1
	mv /tmp/narrative-onnx-linux/lib ./libs/onnx-linux
	rm -Rf /tmp/narrative-onnx-linux

.PHONY: get-lib-darwin
get-lib-darwin:
	mkdir -p ./libs/ /tmp/narrative-onnx-darwin
	wget "https://github.com/microsoft/onnxruntime/releases/download/v1.25.1/onnxruntime-osx-arm64-1.25.1.tgz" -O /tmp/narrative-onnx-darwin/lib.tgz
	tar -xvzf /tmp/narrative-onnx-darwin/lib.tgz  -C /tmp/narrative-onnx-darwin --strip-components=2
	mv /tmp/narrative-onnx-darwin/lib ./libs/onnx-darwin
	rm -Rf /tmp/narrative-onnx-darwin


.PHONY: get-lib-linux-arm64
get-lib-linux-arm64:
	mkdir -p ./libs/ /tmp/narrative-onnx-linux-arm64
	wget "https://github.com/microsoft/onnxruntime/releases/download/v1.25.1/onnxruntime-linux-aarch64-1.25.1.tgz" -O /tmp/narrative-onnx-linux-arm64/lib.tgz
	tar -xvzf /tmp/narrative-onnx-linux-arm64/lib.tgz  -C /tmp/narrative-onnx-linux-arm64 --strip-components=1
	mv /tmp/narrative-onnx-linux-arm64/lib ./libs/onnx-linux-arm64
	rm -Rf /tmp/narrative-onnx-linux-arm64
