package aero

import (
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
)

// Markdown converts the given markdown code to an HTML string.
func Markdown(code string) string {
	return BytesToStringUnsafe(blackfriday.MarkdownCommon(StringToBytesUnsafe(bluemonday.UGCPolicy().Sanitize(code))))
}
