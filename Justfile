set export

DEBUG := "1"

BIN_FILE := "out.bin"
f := ""

run FILE=f:
	go run ./cmd/narrative/main.go {{FILE}}

convert FILE=f:
	go run ./cmd/narrative/main.go --convert {{FILE}}

debug FILE:
	go run ./cmd/debugger/main.go {{FILE}}

merge-dicts:
	go run cmd/dict-merge/main.go dictionary/aux_dict.csv ./narrative-debugger/aux_dict.csv

play bin_file=BIN_FILE:
	ffplay -vn -v quiet -autoexit -f f32le -ar  44100 -showmode 1 {{bin_file}}

playback-continuous:
	./scripts/playback-continuous

check-dict WORD:
    go run ./cmd/check-dict/main.go {{WORD}}

tts WORD:
    go run ./cmd/tts/main.go {{WORD}} && just play

tts-raw PHONEME:
    go run ./cmd/tts-raw/main.go {{PHONEME}} && just play
