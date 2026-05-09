package narrative

type OnnxLib struct {
	remote, os, arch, name, tar_path, filename string
}

// Stores supported onnx libraries
// in a form os->arch->[]lib.
var OnnxLibMap = map[string]map[string][]OnnxLib{
	"linux": {
		"amd64": []OnnxLib{
			{
				remote:   "https://github.com/microsoft/onnxruntime/releases/download/v1.25.1/onnxruntime-linux-x64-1.25.1.tgz",
				os:       "linux",
				arch:     "amd64",
				name:     "onnx-linux-amd64-cpu",
				tar_path: "onnxruntime-linux-x64-1.25.1/lib",
				filename: "libonnxruntime.so",
			},
		},
		"arm64": []OnnxLib{
			{
				remote:   "https://github.com/microsoft/onnxruntime/releases/download/v1.25.1/onnxruntime-linux-aarch64-1.25.1.tgz",
				os:       "linux",
				arch:     "arm64",
				name:     "onnx-linux-arm64-cpu",
				tar_path: "onnxruntime-linux-aarch64-1.25.1/lib",
				filename: "libonnxruntime.so",
			},
		},
	},
	"darwin": {
		"arm64": []OnnxLib{
			{

				remote:   "https://github.com/microsoft/onnxruntime/releases/download/v1.25.1/onnxruntime-osx-arm64-1.25.1.tgz",
				os:       "darwin",
				arch:     "arm64",
				name:     "onnx-darwin-arm64",
				tar_path: "onnxruntime-osx-arm64-1.25.1/lib",
				filename: "libonnxruntime.dylib",
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
			if l.name == name {
				lib = &l
			}
		}
	} else {
		lib = &libs[0]
	}
	return lib
}
