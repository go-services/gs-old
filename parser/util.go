package parser

import (
	"strings"

	"github.com/go-services/annotation"
	"github.com/go-services/code"
)

func findAnnotations(name string, annotations []annotation.Annotation) (found []annotation.Annotation) {
	for _, ann := range annotations {
		if ann.Name == name {
			found = append(found, ann)
		}
	}
	return
}
func getTag(key string, tags code.FieldTags) string {
	tag, _ := tags[key]
	return tag
}
func getParameter(tag string) (name string, required bool) {
	values := strings.Split(strings.Replace(tag, " ", "", -1), ",")
	name = values[0]
	if len(values) > 1 && values[1] == "required" {
		required = true
	}
	return
}
