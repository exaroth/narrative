package english

import "embed"

//go:embed missing.all.zlib language.json weights*.json.zlib aux_dict.csv
var Language embed.FS
