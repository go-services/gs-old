package template

import (
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"

	"github.com/ozgio/strutil"
)

var CustomFunctions = template.FuncMap{
	"lowerFirst":          ToLowerFirst,
	"title":               toTitle,
	"camelCase":           camelCase,
	"httpResponseEncoder": GetHttpResponseEncodeFunction,
	"httpRequestDecoder":  GetHttpRequestDecoderFunction,
}

func GetHttpResponseEncodeFunction(tp string) string {
	return map[string]string{
		"JSON": "jsonEncoder",
		"XML":  "xmlEncoder",
	}[tp]
}

func GetHttpRequestDecoderFunction(tp string) string {
	return map[string]string{
		"JSON": "jsonDecoder",
		"XML":  "xmlDecoder",
		"FORM": "formDecoder",
	}[tp]
}

func ToLowerFirst(text string) string {
	if len(text) > 0 {
		r, size := utf8.DecodeRuneInString(text)
		if r != utf8.RuneError || size > 1 {
			lo := unicode.ToLower(r)
			if lo != r {
				text = string(lo) + text[size:]
			}
		}
	}
	return text
}

func toTitle(text string) string {
	return strings.Title(strutil.ToCamelCase(text))
}

func camelCase(s string) string {
	prep := strings.ReplaceAll(s, "_", " ")
	prep = strings.ReplaceAll(prep, "-", " ")
	return toTitle(strutil.ToCamelCase(prep))
}
