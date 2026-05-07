package narrative

import "path/filepath"

const (
	VOICES_FNAME      = "voices.npz"
	MODEL_LOCAL_FNAME = "kitten.onnx"
)

var (
	TTSModelNano = TTSModel{
		name:   "nano",
		remote: "https://huggingface.co/KittenML/kitten-tts-nano-0.8/resolve/main",
		fname:  "kitten_tts_nano_v0_8.onnx",
	}
	TTSModelMicro = TTSModel{
		name:   "micro",
		remote: "https://huggingface.co/KittenML/kitten-tts-micro-0.8/resolve/main/kitten_tts_micro_v0_8.onnx",
		fname:  "kitten_tts_micro_v0_8.onnx",
	}
	TTSModelMini = TTSModel{
		name:   "mini",
		remote: "https://huggingface.co/KittenML/kitten-tts-mini-0.8/resolve/main/kitten_tts_mini_v0_8.onnx",
		fname:  "kitten_tts_mini_v0_8.onnx",
	}
)

// Represents model settings as saved in data config.
type TTSModel struct {
	name, fname, remote string
}

func (t *TTSModel) GetModelDir(base string) string {
	return filepath.Join(base, t.name)
}

func (t *TTSModel) GetModelPath(base string) string {
	return filepath.Join(t.GetModelDir(base), MODEL_LOCAL_FNAME)
}

func (t *TTSModel) GetVoicesPath(base string) string {
	return filepath.Join(t.GetModelDir(base), VOICES_FNAME)
}
