package main

import (
	"fmt"
	"log"

	"github.com/sbinet/npyio/npz"
	ort "github.com/yalue/onnxruntime_go"
)

var VOICE_MAP map[string]string = map[string]string{
	"Bella":  "expr-voice-2-f.npy",
	"Jasper": "expr-voice-2-m.npy",
	"Luna":   "expr-voice-3-f.npy",
	"Bruno":  "expr-voice-3-m.npy",
	"Rosie":  "expr-voice-4-f.npy",
	"Hugo":   "expr-voice-4-m.npy",
	"Kiki":   "expr-voice-5-f.npy",
	"Leo":    "expr-voice-5-m.npy",
}

const DEFAULT_VOICE = "Hugo"
const DEFAULT_SPEED float32 = 0.9

type vMat [400][256]float32

func (v *vMat) Load(flat []float32) {
	if len(flat) != 102400 {
		panic("Invalid voice matrix length, expected 102400")
	}
	for i := range 400 {
		v[i] = ([256]float32)(flat[i*256 : (i+1)*256])
	}

}

type Kitten struct {
	voice   *vMat
	session *ort.DynamicAdvancedSession
}

func (k *Kitten) Deinit() {
	k.session.Destroy()
	ort.DestroyEnvironment()
}

func (k *Kitten) createTensors(sentence []int64) (
	input *ort.Tensor[int64],
	voice *ort.Tensor[float32],
	speed *ort.Tensor[float32],
	err error,
) {
	sentence = append([]int64{0}, sentence...)
	sentence = append(sentence, 10)
	sentence = append(sentence, 0)

	inputShape := ort.NewShape(1, int64(len(sentence)))
	inputTensor, err := ort.NewTensor(inputShape, sentence)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Error creating input tensor: %+v\n, %w", sentence, err)
	}
	voiceData := k.voice[1][:]

	voiceShape := ort.NewShape(1, 256)
	voiceTensor, err := ort.NewTensor(voiceShape, voiceData)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Error creating voice tensor: %w", err)
	}

	speedShape := ort.NewShape(1)
	speedTensor, err := ort.NewTensor(speedShape, []float32{DEFAULT_SPEED})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Error creating speed tensor: %w", err)
	}
	return inputTensor, voiceTensor, speedTensor, nil
}

func (k *Kitten) RunInference(sentence []int64) ([]float32, error) {

	inputTensor, voiceTensor, speedTensor, err := k.createTensors(sentence)
	if err != nil {
		return nil, err
	}
	defer inputTensor.Destroy()
	defer voiceTensor.Destroy()
	defer speedTensor.Destroy()

	outputs := []ort.Value{nil, nil}
	defer func() {
		for _, o := range outputs {
			o.Destroy()
		}
	}()
	err = k.session.Run(
		[]ort.Value{inputTensor, voiceTensor, speedTensor},
		outputs,
	)
	if err != nil {
		return nil, fmt.Errorf("Error running inference: %w", err)
	}
	waveform := outputs[0].(*ort.Tensor[float32]).GetData()
	return waveform, nil
}

func NewKitten(voice_name *string) *Kitten {
	ort.SetSharedLibraryPath("/home/exaroth/Projects/narrative/onnx_libs/lib/libonnxruntime.so")

	err := ort.InitializeEnvironment()
	if err != nil {
		log.Fatalf("Could intialize onnx runtime: %+v", err)
	}
	var voice_dtf string
	if voice_name != nil {
		v, ok := VOICE_MAP[*voice_name]
		if !ok {
			log.Fatalf("Invalid voice id %s provided", voice_name)
		}
		voice_dtf = v
	} else {
		voice_dtf = VOICE_MAP[DEFAULT_VOICE]
	}

	f, err := npz.Open("./kitten/voices.npz")
	if err != nil {
		log.Fatalf("Could not open npz file: %+v", err)
	}
	defer f.Close()

	var f0 []float32

	err = f.Read(voice_dtf, &f0)
	if err != nil {
		log.Fatalf("Could not read value from npz file: %+v", err)
	}
	var voice vMat
	voice.Load(f0)

	session, err := ort.NewDynamicAdvancedSession(
		"./kitten/kitten.onnx",
		[]string{"input_ids", "style", "speed"},
		[]string{"waveform", "duration"},
		nil,
	)
	if err != nil {
		log.Fatalf("Error intializing session %+v", err)
	}

	return &Kitten{
		voice:   &voice,
		session: session,
	}

}
