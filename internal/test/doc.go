package test

// NestedStruct is a nested structure with documentation.
type NestedStruct struct {
	// Baz is a field within a field.
	Baz string
}

// DocStruct is a structure with documentation.
type DocStruct struct {
	// Foo is a field.
	Foo string

	// Bar is also a field.
	Bar *NestedStruct

	// Tagged is renamed by its json tag.
	Tagged string `json:"tagged"`

	// Renamed is renamed by its openapi tag.
	Renamed string `json:"renamed_json" openapi:"renamed"`

	Trailing string // Trailing is documented on the same line.

	// Both share one comment.
	Both1, Both2 string
}

// DocRequest is a request whose parameters are documented.
type DocRequest struct {
	// Since is an RFC 3339 timestamp. When given, the response includes
	// deleted items.
	Since string `json:"since" openapi:",in=query"`

	// ID identifies the item.
	ID string `json:"id" openapi:",in=path"`
}
