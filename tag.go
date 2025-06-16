package arrest

import (
	"reflect"
	"strings"
)

type JSONTag string

func (tag JSONTag) Parts() []string {
	return strings.Split(string(tag), ",")
}

func (tag JSONTag) IsIgnored() bool {
	return len(tag.Parts()) > 0 && tag.Name() == "-"
}

func (tag JSONTag) HasName() bool {
	return len(tag.Parts()) > 0 && tag.Name() != "-" && tag.Name() != ""
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

func (tag OpenAPITag) IsIgnored() bool {
	return len(tag.Parts()) > 0 && tag.Name() == "-"
}

func (tag OpenAPITag) HasName() bool {
	return len(tag.Parts()) > 0 && tag.Name() != "-" && tag.Name() != ""
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
		if len(pair) == 2 {
			props[strings.TrimSpace(pair[0])] = strings.TrimSpace(pair[1])
			continue
		}
		props[strings.TrimSpace(part)] = "true"
	}
	return props
}

type TagInfo struct {
	jsonTag    JSONTag
	openAPITag OpenAPITag
}

func NewTagInfo(tag reflect.StructTag) *TagInfo {
	jsonTag := JSONTag(tag.Get("json"))
	openApiTag := OpenAPITag(tag.Get("openapi"))

	return &TagInfo{
		jsonTag:    jsonTag,
		openAPITag: openApiTag,
	}
}

func (info *TagInfo) IsIgnored() bool {
	return info.jsonTag.IsIgnored() || info.openAPITag.IsIgnored()
}

func (info *TagInfo) HasName() bool {
	return info.openAPITag.HasName() || info.jsonTag.HasName()
}

func (info *TagInfo) Name() string {
	switch {
	case info.openAPITag.HasName():
		return info.openAPITag.Name()
	case info.jsonTag.HasName():
		return info.jsonTag.Name()
	}
	return ""
}

func (info *TagInfo) Props() map[string]string {
	return info.openAPITag.Props()
}

func (info *TagInfo) ReplacementType() string {
	return info.Props()["type"]
}

func (info *TagInfo) RefName() string {
	return info.Props()["refName"]
}

func (info *TagInfo) ElemRefName() string {
	return info.Props()["elemRefName"]
}

func (info *TagInfo) HasIn() bool {
	return info.Props()["in"] != ""
}

func (into *TagInfo) In() string {
	return into.Props()["in"]
}

func (info *TagInfo) Recursive() bool { return info.Props()["recursive"] == "true" }
