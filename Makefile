.PHONY: build
build:
	# go build -o ./build/narrative ./cmd/narrative/main.go
	go build -o narrative ./cmd/narrative/main.go

.PHONY: build-debugger
build-debugger:
	go build -o debugger ./cmd/debugger/main.go

.PHONY: run
run:
	DEBUG=1 go run ./cmd/narrative/main.go

# Setup test env
.PHONY: test-setup
test-setup:
	go get github.com/kyoh86/richgo
	go get github.com/jstemmer/go-junit-report

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
