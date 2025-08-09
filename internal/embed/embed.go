package embed

import (
	"embed"
)

//go:embed all:static
var StaticFiles embed.FS
