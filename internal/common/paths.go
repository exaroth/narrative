package common

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
)

// path constants.
const (
	// name of dir used for storing tts model data.
	DEFAULT_MODEL_DIR_NAME = "models"
	// name of dir used for storing onnxruntime libs.
	DEFAULT_LIB_DIR_NAME = "lib"
	// default config filename.
	DEFAULT_CONFIG_FNAME = "config.yaml"
	// default data config filename.
	DEFAULT_DATA_CONFIG_FNAME = "config.json"
	// default sources dir name
	DEFAULT_SOURCES_DIR_NAME = "sources"
	// default dir name for ~/.local/share and ~/.config.
	DEFAULT_CONFIG_DIR_NAME = "narrative"
)

func init() {
	switch runtime.GOOS {
	case "linux", "freebsd":
		getPathsXdg()
	case "darwin":
		getPathsDarwin()
	default:
		panic("Unsupported system")
	}
}

var (
	localConfig string
	localData   string
	localCache  string
)

// Create basic directory structure for narrative.
func InitDirectoryStructure() *NarrativePaths {
	paths := InitPaths()
	MakePath(paths.ConfigDir)
	MakePath(paths.DataDir)
	MakePath(paths.ModelPath)
	MakePath(paths.LibPath)
	MakePath(paths.SourcesPath)
	return paths
}

// Contains all filesystem paths used by narrative.
type NarrativePaths struct {
	ConfigDir, DataDir, CacheDir, ConfigPath, ModelPath,
	LibPath, DataConfigPath, SourcesPath string
}

// Check if narrative requires initialization
// on startup.
func (p *NarrativePaths) RequiresInit() bool {
	for _, path := range [2]string{
		p.DataConfigPath,
		p.ConfigPath,
	} {
		if _, err := os.Stat(path); err != nil {
			return true
		}
	}
	for _, path := range [2]string{
		p.ModelPath,
		p.LibPath,
	} {
		if CheckDirEmpty(path) {
			return true
		}
	}
	return false
}

func InitPaths() *NarrativePaths {
	return &NarrativePaths{
		ConfigDir:      LocalConfig(),
		DataDir:        LocalData(),
		CacheDir:       LocalCache(),
		ConfigPath:     LocalConfig(DEFAULT_CONFIG_FNAME),
		ModelPath:      LocalData(DEFAULT_MODEL_DIR_NAME),
		LibPath:        LocalData(DEFAULT_LIB_DIR_NAME),
		DataConfigPath: LocalData(DEFAULT_DATA_CONFIG_FNAME),
		SourcesPath:    LocalData(DEFAULT_SOURCES_DIR_NAME),
	}
}

func getPathsXdg() {

	if os.Getenv("XDG_CONFIG_HOME") != "" {
		localConfig = os.Getenv("XDG_CONFIG_HOME")
	} else {
		localConfig = filepath.Join(os.Getenv("HOME"), ".config")
	}
	if os.Getenv("XDG_DATA_HOME") != "" {
		localData = os.Getenv("XDG_DATA_HOME")
	} else {
		localData = filepath.Join(os.Getenv("HOME"), ".local/share")
	}
	if os.Getenv("XDG_CACHE_HOME") != "" {
		localCache = os.Getenv("XDG_CACHE_HOME")
	} else {
		localCache = filepath.Join(os.Getenv("HOME"), ".cache")
	}
}

func getPathsDarwin() {
	localConfig = os.Getenv("HOME") + "/Library/Application Support"
	localData = os.Getenv("HOME") + "/Library/Application Support"
	localCache = os.Getenv("HOME") + "/Library/Caches"
}

// LocalConfig returns the local user configuration path, with optional
// path components added to the end for vendor/application-specific settings.
func LocalConfig(dirs ...string) string {
	if len(dirs) == 0 {
		return filepath.Join(localConfig, DEFAULT_CONFIG_DIR_NAME)
	}
	return filepath.Join(localConfig, DEFAULT_CONFIG_DIR_NAME, filepath.Join(dirs...))
}

// LocalData stores any additional data needed by narrative,
// eg. models, libraries etc. For MacOS its same as config dir.
func LocalData(dirs ...string) string {
	if len(dirs) == 0 {
		return filepath.Join(localData, DEFAULT_CONFIG_DIR_NAME)
	}

	return filepath.Join(localData, DEFAULT_CONFIG_DIR_NAME, filepath.Join(dirs...))
}

// LocalCache returns the local user cache folder, with optional path
// components added to the end for vendor/application-specific settings.
func LocalCache(folder ...string) string {
	if len(folder) == 0 {
		return filepath.Join(localCache, DEFAULT_CONFIG_DIR_NAME)
	}

	return filepath.Join(localCache, DEFAULT_CONFIG_DIR_NAME, filepath.Join(folder...))
}

func MakePath(path string) error {
	return os.MkdirAll(path, os.FileMode(0755))
}

func CheckDirEmpty(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return true
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true
	}
	return false
}
