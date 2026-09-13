package arrest

import (
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

// Response provides DSL methods for creating OpenAPI responses. The
// libopenapi response underneath is available through OpenAPIResponse.
type Response struct {
	state any // *v3.Response; see state.go

	ErrHelper
}

// Description sets the description of the response.
//
//go:noinline
func (r *Response) Description(description string) *Response {
	responseOf(r).Description = description
	return r
}

// Header adds a header to the response.
//
//go:noinline
func (r *Response) Header(name string, m *Model, mods ...func(h *Header)) *Response {
	res := responseOf(r)
	if res.Headers == nil {
		res.Headers = orderedmap.New[string, *v3.Header]()
	}

	hdr := &v3.Header{}
	res.Headers.Set(name, hdr)

	m.AddHandler(r)
	hdr.Schema = schemaOf(m)

	if len(mods) > 0 {
		h := &Header{state: hdr}
		for _, mod := range mods {
			mod(h)
		}
	}

	return r
}

// Content adds a content type to the response.
//
//go:noinline
func (r *Response) Content(code string, m *Model) *Response {
	res := responseOf(r)
	if res.Content == nil {
		res.Content = orderedmap.New[string, *v3.MediaType]()
	}

	m.AddHandler(r)
	res.Content.Set(code, &v3.MediaType{Schema: schemaOf(m)})
	return r
}

// ContentMediaType is used to specify when the response is a raw binary type.
//
//go:noinline
func (r *Response) ContentMediaType(mediaType string) *Response {
	res := responseOf(r)
	if res.Content == nil {
		res.Content = orderedmap.New[string, *v3.MediaType]()
	}

	res.Content.Set(mediaType, &v3.MediaType{})

	return r
}
