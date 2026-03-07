package storage

import "strings"

// Category classifies uploaded files.
type Category string

const (
	CategoryImage    Category = "image"
	CategoryVideo    Category = "video"
	CategoryAudio    Category = "audio"
	CategoryPDF      Category = "pdf"
	CategoryMarkdown Category = "markdown"
	CategoryFile     Category = "file"
)

// NormalizeCategory maps an arbitrary string to a known Category.
func NormalizeCategory(raw string) Category {
	switch Category(strings.ToLower(strings.TrimSpace(raw))) {
	case CategoryImage:
		return CategoryImage
	case CategoryVideo:
		return CategoryVideo
	case CategoryAudio:
		return CategoryAudio
	case CategoryPDF:
		return CategoryPDF
	case CategoryMarkdown:
		return CategoryMarkdown
	default:
		return CategoryFile
	}
}

// IsImageLike reports whether the category represents visual media.
func (c Category) IsImageLike() bool {
	return c == CategoryImage
}

// URLResolver maps a VireFS key to a publicly accessible URL.
// It is constructed once at startup via NewURLResolver and applies
// schema.Resolve internally, so callers just pass the flat key stored
// in the database.
type URLResolver func(key string) string

// KeyGenerator produces a flat filename key (no directory prefix).
// VireFS Schema handles the directory routing transparently.
type KeyGenerator interface {
	GenerateKey(category Category, userID uint, originalFilename string) (string, error)
}

// TrimLeadingSlash removes a leading "/" to convert a virtual path to a
// VireFS key.
func TrimLeadingSlash(p string) string {
	return strings.TrimPrefix(p, "/")
}
