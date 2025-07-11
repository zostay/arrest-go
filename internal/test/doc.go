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
}
