package main

import (
	"fmt"
	"log"

	"github.com/sbinet/npyio/npz"
	ort "github.com/yalue/onnxruntime_go"
)

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
	inputShape := ort.NewShape(1, 24)
	inputTensor, err := ort.NewTensor(inputShape, sentence)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Error creating input tensor: %+v\n, %w", sentence, err)
	}
	voiceData := k.voice[24][:]
	// fmt.Println(voiceData)

	voiceShape := ort.NewShape(1, 256)
	voiceTensor, err := ort.NewTensor(voiceShape, voiceData)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Error creating voice tensor: %w", err)
	}

	speedShape := ort.NewShape(1)
	speedTensor, err := ort.NewTensor(speedShape, []float32{1.0})
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

	outputShape := ort.NewShape(78000)
	outputTensor, err := ort.NewEmptyTensor[float32](outputShape)
	if err != nil {
		return nil, fmt.Errorf("Error creating output tensor: %w", err)

	}
	defer outputTensor.Destroy()

	err = k.session.Run(
		[]ort.Value{inputTensor, voiceTensor, speedTensor},
		[]ort.Value{outputTensor},
	)
	if err != nil {
		return nil, fmt.Errorf("Error running inference: %w", err)
	}
	return outputTensor.GetData(), nil
}

func NewKitten() *Kitten {
	ort.SetSharedLibraryPath("/home/exaroth/Projects/narrative/onnx_libs/lib/libonnxruntime.so")

	err := ort.InitializeEnvironment()
	if err != nil {
		log.Fatalf("Could intialize onnx runtime: %+v", err)
	}

	f, err := npz.Open("./kitten/voices.npz")
	if err != nil {
		log.Fatalf("Could not open npz file: %+v", err)
	}
	defer f.Close()

	var f0 []float32

	err = f.Read("expr-voice-2-m.npy", &f0)
	if err != nil {
		log.Fatalf("Could not read value from npz file: %+v", err)
	}
	var voice vMat
	voice.Load(f0)

	session, err := ort.NewDynamicAdvancedSession(
		"./kitten/kitten.onnx",
		[]string{"input_ids", "style", "speed"},
		[]string{"waveform"},
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
