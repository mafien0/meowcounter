// Package assets embeds assets into an application
package assets

import (
	"embed"
)

//go:embed *.png
var FS embed.FS
