
.PHONY: build
build:
	go build -o ./build/narrative ./cmd/narrative/main.go

.PHONY: run
run:
	go run ./cmd/narrative/main.go

.PHONY: play
play:
	ffplay  -vn -v quiet -autoexit -f f32le -ar  44100 -showmode 1 out.bin

.PHONY: debugger
debugger:
	go run ./cmd/debugger/main.go ./dump/kafka-on-the-shore.txt


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
