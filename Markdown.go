package aero

import (
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
)

var strictPolicy = bluemonday.StrictPolicy()

// Markdown converts the given markdown code to an HTML string.
func Markdown(code string) string {
	codeBytes := StringToBytesUnsafe(code)
	codeBytes = strictPolicy.SanitizeBytes(codeBytes)
	codeBytes = blackfriday.MarkdownCommon(codeBytes)
	return BytesToStringUnsafe(codeBytes)
}
