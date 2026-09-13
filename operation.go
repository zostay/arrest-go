package arrest

import (
	"fmt"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

// Operation provides DSL methods for creating OpenAPI operations. The
// libopenapi operation underneath is available through OpenAPIOperation.
type Operation struct {
	state any // *v3.Operation; see state.go

	ErrHelper
}

// RequestBody sets the request body for the operation.
//
//go:noinline
func (o *Operation) RequestBody(mt string, model *Model) *Operation {
	sp := schemaOf(model)
	if sp == nil {
		return withErr(o, fmt.Errorf("model must be initialized"))
	}

	o.AddHandler(model)

	op := opOf(o)
	if op.RequestBody == nil {
		op.RequestBody = &v3.RequestBody{}
	}

	if op.RequestBody.Content == nil {
		op.RequestBody.Content = orderedmap.New[string, *v3.MediaType]()
	}

	mts := op.RequestBody.Content
	mts.Set(mt, &v3.MediaType{Schema: sp})

	return o
}

// Description sets the description for the operation.
//
//go:noinline
func (o *Operation) Description(description string) *Operation {
	opOf(o).Description = description
	return o
}

// Summary sets the summary for the operation.
//
//go:noinline
func (o *Operation) Summary(summary string) *Operation {
	opOf(o).Summary = summary
	return o
}

// OperationID sets the operation ID for the operation.
//
//go:noinline
func (o *Operation) OperationID(id string) *Operation {
	opOf(o).OperationId = id
	return o
}

// Tags adds tags to the operation.
//
//go:noinline
func (o *Operation) Tags(tags ...string) *Operation {
	op := opOf(o)
	op.Tags = append(op.Tags, tags...)
	return o
}

// Deprecated marks the operation as deprecated.
//
//go:noinline
func (o *Operation) Deprecated() *Operation {
	deprecated := true
	opOf(o).Deprecated = &deprecated
	return o
}

// Parameters adds parameters to the operation.
//
//go:noinline
func (o *Operation) Parameters(ps *Parameters) *Operation {
	op := opOf(o)
	if op.Parameters == nil {
		op.Parameters = []*v3.Parameter{}
	}

	o.AddHandler(ps)

	for _, p := range ps.Parameters {
		op.Parameters = append(op.Parameters, paramOf(p))
	}

	return o
}

// Response adds a response to the operation.
//
//go:noinline
func (o *Operation) Response(code string, cb func(r *Response)) *Operation {
	op := opOf(o)
	if op.Responses == nil {
		op.Responses = &v3.Responses{}
	}

	if op.Responses.Codes == nil {
		op.Responses.Codes = orderedmap.New[string, *v3.Response]()
	}

	codes := op.Responses.Codes
	if _, hasCode := codes.Get(code); !hasCode {
		codes.Set(code, &v3.Response{})
	}

	res := &Response{state: codes.GetOrZero(code)}
	o.AddHandler(res)

	cb(res)

	return o
}

// HasResponses reports whether any response has been added to the operation.
//
//go:noinline
func (o *Operation) HasResponses() bool {
	op := opOf(o)
	return op.Responses != nil && op.Responses.Codes != nil && op.Responses.Codes.Len() > 0
}

// SecurityRequirement configures the security scopes for this operation. The key in
// the map is the security scheme name and the value is the list of scopes.
//
//go:noinline
func (o *Operation) SecurityRequirement(reqs map[string][]string) *Operation {
	op := opOf(o)
	if op.Security == nil {
		op.Security = []*base.SecurityRequirement{}
	}

	op.Security = append(op.Security, &base.SecurityRequirement{
		Requirements: orderedmap.ToOrderedMap(reqs),
	})

	return o
}
