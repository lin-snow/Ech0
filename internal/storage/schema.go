package storage

import virefs "github.com/lin-snow/VireFS"

// NewFileSchema builds a VireFS Schema that routes files into
// subdirectories by extension. Plug it into VireFS via
// WithLocalKeyFunc(schema.Resolve) or WithObjectKeyFunc(schema.Resolve).
func NewFileSchema() *virefs.Schema {
	return virefs.NewSchema(
		virefs.RouteByExt("images/", ".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg", ".avif"),
		virefs.RouteByExt("audios/", ".mp3", ".flac", ".wav", ".m4a", ".ogg"),
		virefs.RouteByExt("videos/", ".mp4", ".avi", ".mkv", ".webm"),
		virefs.RouteByExt("documents/", ".pdf", ".doc", ".docx"),
		virefs.DefaultRoute("files/"),
	)
}
