package planmark

import (
	"embed"
	"io/fs"
)

//go:embed all:web/dist
var WebAssets embed.FS

// GetSubFS returns a sub-filesystem rooted at web/dist
func GetSubFS() fs.FS {
	sub, err := fs.Sub(WebAssets, "web/dist")
	if err != nil {
		return WebAssets
	}
	return sub
}
