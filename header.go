package arrest

import v3 "github.com/pb33f/libopenapi/datamodel/high/v3"

// Header provides DSL methods for creating OpenAPI headers.
type Header struct {
	Header *v3.Header
}

// Description sets the description of the header.
func (h *Header) Description(description string) *Header {
	h.Header.Description = description
	return h
}
