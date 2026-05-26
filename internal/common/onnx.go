package common

type OnnxLib struct {
	Remote, Os, Arch, Name, TarPath, Filename string
}

// Stores supported onnx libraries
// in a form os->arch->[]lib.
var OnnxLibMap = map[string]map[string][]OnnxLib{
	"linux": {
		"amd64": []OnnxLib{
			{
				Remote:   "https://github.com/microsoft/onnxruntime/releases/download/v1.25.1/onnxruntime-linux-x64-1.25.1.tgz",
				Os:       "linux",
				Arch:     "amd64",
				Name:     "onnx-linux-amd64-cpu",
				TarPath:  "onnxruntime-linux-x64-1.25.1/lib",
				Filename: "libonnxruntime.so",
			},
			{
				Remote:   "https://github.com/microsoft/onnxruntime/releases/download/v1.25.1/onnxruntime-linux-x64-gpu_cuda13-1.25.1.tgz",
				Os:       "linux",
				Arch:     "amd64",
				Name:     "onnx-linux-amd64-cuda",
				TarPath:  "onnxruntime-linux-x64-gpu-1.25.1/lib",
				Filename: "libonnxruntime.so",
			},
			{
				Remote:   "https://github.com/microsoft/onnxruntime/releases/download/v1.25.1/onnxruntime-linux-x64-gpu-1.25.1.tgz",
				Os:       "linux",
				Arch:     "amd64",
				Name:     "onnx-linux-amd64-gpu",
				TarPath:  "onnxruntime-linux-x64-gpu-1.25.1/lib",
				Filename: "libonnxruntime.so",
			},
		},
		"arm64": []OnnxLib{
			{
				Remote:   "https://github.com/microsoft/onnxruntime/releases/download/v1.25.1/onnxruntime-linux-aarch64-1.25.1.tgz",
				Os:       "linux",
				Arch:     "arm64",
				Name:     "onnx-linux-arm64-cpu",
				TarPath:  "onnxruntime-linux-aarch64-1.25.1/lib",
				Filename: "libonnxruntime.so",
			},
		},
	},
	"darwin": {
		"arm64": []OnnxLib{
			{

				Remote:   "https://github.com/microsoft/onnxruntime/releases/download/v1.25.1/onnxruntime-osx-arm64-1.25.1.tgz",
				Os:       "darwin",
				Arch:     "arm64",
				Name:     "onnx-darwin-arm64",
				TarPath:  "onnxruntime-osx-arm64-1.25.1/lib",
				Filename: "libonnxruntime.dylib",
			},
		},
	},
}

// Retrieve onnx lib data based on arch and os.
func GetOnnxLib(os, arch, name string) *OnnxLib {
	var lib *OnnxLib
	var libs []OnnxLib
	if os_libs, ok := OnnxLibMap[os]; !ok {
		return nil
	} else {
		if arch_libs, ok := os_libs[arch]; !ok {
			return nil
		} else {
			libs = arch_libs
		}
	}

	if len(libs) == 0 {
		return nil
	}
	if len(name) > 0 {
		for _, l := range libs {
			if l.Name == name {
				lib = &l
			}
		}
	} else {
		lib = &libs[0]
	}
	return lib
}
