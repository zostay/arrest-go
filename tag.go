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

// IsDiscriminator returns true if this field is marked as a discriminator
func (info *TagInfo) IsDiscriminator() bool {
	_, exists := info.Props()["discriminator"]
	return exists
}

// GetDefaultMapping returns the default mapping value for a discriminator
func (info *TagInfo) GetDefaultMapping() string {
	return info.Props()["defaultMapping"]
}

// GetPolymorphType returns the polymorphic composition type (oneOf, anyOf, allOf)
func (info *TagInfo) GetPolymorphType() string {
	props := info.Props()
	if _, exists := props["oneOf"]; exists {
		return "oneOf"
	}
	if _, exists := props["anyOf"]; exists {
		return "anyOf"
	}
	if _, exists := props["allOf"]; exists {
		return "allOf"
	}
	return ""
}

// GetMapping returns the discriminator mapping alias for this field
func (info *TagInfo) GetMapping() string {
	return info.Props()["mapping"]
}

// IsInline returns true if this field should be inlined (from json tag)
func (info *TagInfo) IsInline() bool {
	parts := info.jsonTag.Parts()
	for _, part := range parts[1:] {
		if strings.TrimSpace(part) == "inline" {
			return true
		}
	}
	return false
}

// IsOmitEmpty returns true if this field has omitempty (from json tag)
func (info *TagInfo) IsOmitEmpty() bool {
	parts := info.jsonTag.Parts()
	for _, part := range parts[1:] {
		if strings.TrimSpace(part) == "omitempty" {
			return true
		}
	}
	return false
}
