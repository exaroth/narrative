
.PHONY: run
run:
	go run main.go tokenizer.go kitten.go

.PHONY: play
play:
	ffplay -f f32le -ar  44100 -showmode 1 out.bin
