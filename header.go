package arrest

// Header provides DSL methods for creating OpenAPI headers. The libopenapi
// header underneath is available through OpenAPIHeader.
type Header struct {
	state any // *v3.Header; see state.go
}

// Description sets the description of the header.
//
//go:noinline
func (h *Header) Description(description string) *Header {
	headerOf(h).Description = description
	return h
}
