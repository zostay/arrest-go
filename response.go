package arrest

import (
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

type Response struct {
	Response *v3.Response

	ErrHelper
}

func (r *Response) Description(description string) *Response {
	r.Response.Description = description
	return r
}

func (r *Response) Header(name string, m *Model, mods ...func(h *Header)) *Response {
	if r.Response.Headers == nil {
		r.Response.Headers = orderedmap.New[string, *v3.Header]()
	}

	hdr := &v3.Header{}
	r.Response.Headers.Set(name, hdr)

	m.AddHandler(m)
	hdr.Schema = m.SchemaProxy

	if len(mods) > 0 {
		h := &Header{Header: hdr}
		for _, mod := range mods {
			mod(h)
		}
	}

	return r
}

func (r *Response) Content(code string, m *Model) *Response {
	if r.Response.Content == nil {
		r.Response.Content = orderedmap.New[string, *v3.MediaType]()
	}

	m.AddHandler(m)
	r.Response.Content.Set(code, &v3.MediaType{Schema: m.SchemaProxy})
	return r
}
