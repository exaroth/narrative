set export

DEBUG := "1"

BIN_FILE := "out.bin"
f := ""

run FILE=f:
	go run ./cmd/narrative/main.go {{FILE}}

convert FILE=f:
	go run ./cmd/narrative/main.go --convert {{FILE}}

debug FILE:
	go run ./cmd-aux/debugger/main.go {{FILE}}

merge-dicts:
	go run cmd-aux/dict-merge/main.go dictionary/aux_dict.csv ./narrative-debugger/aux_dict.csv

check-dict WORD:
    go run ./cmd-aux/check-dict/main.go {{WORD}}

tts WORD:
    go run ./cmd-aux/tts/main.go {{WORD}}

tts-raw PHONEME:
    go run ./cmd-aux/tts-raw/main.go {{PHONEME}}

play bin_file=BIN_FILE:
	ffplay -vn -v quiet -autoexit -f f32le -ar  44100 -showmode 1 {{bin_file}}

playback-continuous:
	./scripts/playback-continuous
