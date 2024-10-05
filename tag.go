package arrest

import (
	"reflect"
	"strings"
)

type JSONTag string

func (tag JSONTag) Parts() []string {
	return strings.Split(string(tag), ",")
}

func (tag JSONTag) HasName() bool {
	return len(tag.Parts()) > 0 && tag.Name() != "-"
}

func (tag JSONTag) Name() string {
	parts := tag.Parts()
	if len(parts) == 0 {
		return ""
	}
	return strings.TrimSpace(parts[0])
}

type OpenAPITag string

func (tag OpenAPITag) Parts() []string {
	return strings.Split(string(tag), ",")
}

func (tag OpenAPITag) HasName() bool {
	return len(tag.Parts()) > 0 && tag.Name() != "-"
}

func (tag OpenAPITag) Name() string {
	parts := tag.Parts()
	if len(parts) == 0 {
		return ""
	}
	return strings.TrimSpace(parts[0])
}

func (tag OpenAPITag) Props() map[string]string {
	props := make(map[string]string)
	parts := tag.Parts()
	for _, part := range parts[1:] {
		pair := strings.Split(part, "=")
		if len(pair) != 2 {
			continue
		}
		props[strings.TrimSpace(pair[0])] = strings.TrimSpace(pair[1])
	}
	return props
}

type TagInfo struct {
	Name  string
	Props map[string]string
}

func NewTagInfo(tag reflect.StructTag) *TagInfo {
	jsonTag := JSONTag(tag.Get("json"))
	openApiTag := OpenAPITag(tag.Get("openapi"))

	name := "-"
	switch {
	case openApiTag.HasName():
		name = openApiTag.Name()
	case jsonTag.HasName():
		name = jsonTag.Name()
	}

	return &TagInfo{
		Name:  name,
		Props: openApiTag.Props(),
	}
}

func (info *TagInfo) HasName() bool {
	return info.Name != "-" && info.Name != ""
}

func (info *TagInfo) ReplacementType() string {
	return info.Props["type"]
}
