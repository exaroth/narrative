
# wget "https://huggingface.co/KittenML/kitten-tts-micro-0.8/resolve/main/kitten_tts_micro_v0_8.onnx?download=true" -O kitten_tts_micro_v0_8.onnx

# wget "https://huggingface.co/KittenML/kitten-tts-mini-0.8/resolve/main/kitten_tts_mini_v0_8.onnx?download=true" -O kitten_tts_mini_v0_8.onnx


# wget "https://huggingface.co/KittenML/kitten-tts-nano-0.8/resolve/main/kitten_tts_nano_v0_8.onnx?download=true" -O kitten_tts_nano_v0_8.onnx

# wget "https://huggingface.co/KittenML/kitten-tts-nano-0.8-int8/resolve/main/kitten_tts_nano_v0_8.onnx?download=true" -O kitten_tts_nano_v0_8_int8.onnx

.PHONY: run
run:
	go run main.go tokenizer.go kitten.go "Test input"

.PHONY: play
play:
	ffplay -f f32le -ar  44100 -showmode 1 out.bin
