set export

DEBUG := "1"

BIN_FILE := "out.bin"

run:
	go run ./cmd/narrative/main.go

run-file FILE:
	go run ./cmd/narrative/main.go {{FILE}}

debug FILE:
	go run ./cmd/debugger/main.go {{FILE}}

merge-dicts:
	go run cmd/dict-merge/main.go dictionary/aux_dict.csv ./narrative-debugger/aux_dict.csv

play bin_file=BIN_FILE:
	ffplay -vn -v quiet -autoexit -f f32le -ar  44100 -showmode 1 {{bin_file}}

playback-continuous:
	./scripts/playback-continuous

check-dict WORD:
	./scripts/check-dict {{WORD}}

tts WORD:
	./scripts/tts {{WORD}}

tts-raw PHONEME:
	./scripts/tts {{PHONEME}}
