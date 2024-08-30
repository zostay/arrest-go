package arrest

import v3 "github.com/pb33f/libopenapi/datamodel/high/v3"

type Header struct {
	Header *v3.Header
}

func (h *Header) Description(description string) *Header {
	h.Header.Description = description
	return h
}
