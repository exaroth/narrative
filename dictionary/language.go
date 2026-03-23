package dictionary

import "embed"

//go:embed missing* language.json weights*.json.zlib
var Language embed.FS
