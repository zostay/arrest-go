package arrest

import (
	"fmt"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

// Operation provides DSL methods for creating OpenAPI operations.
type Operation struct {
	Operation *v3.Operation

	ErrHelper
}

// RequestBody sets the request body for the operation.
func (o *Operation) RequestBody(mt string, model *Model) *Operation {
	if model.SchemaProxy == nil {
		return withErr(o, fmt.Errorf("model must be initialized"))
	}

	o.AddHandler(model)

	if o.Operation.RequestBody == nil {
		o.Operation.RequestBody = &v3.RequestBody{}
	}

	if o.Operation.RequestBody.Content == nil {
		o.Operation.RequestBody.Content = orderedmap.New[string, *v3.MediaType]()
	}

	mts := o.Operation.RequestBody.Content
	mts.Set(mt, &v3.MediaType{Schema: model.SchemaProxy})

	return o
}

// Description sets the description for the operation.
func (o *Operation) Description(description string) *Operation {
	o.Operation.Description = description
	return o
}

// Summary sets the summary for the operation.
func (o *Operation) Summary(summary string) *Operation {
	o.Operation.Summary = summary
	return o
}

// OperationID sets the operation ID for the operation.
func (o *Operation) OperationID(id string) *Operation {
	o.Operation.OperationId = id
	return o
}

// Tags adds tags to the operation.
func (o *Operation) Tags(tags ...string) *Operation {
	o.Operation.Tags = append(o.Operation.Tags, tags...)
	return o
}

// Deprecated marks the operation as deprecated.
func (o *Operation) Deprecated() *Operation {
	deprecated := true
	o.Operation.Deprecated = &deprecated
	return o
}

// Parameters adds parameters to the operation.
func (o *Operation) Parameters(ps *Parameters) *Operation {
	if o.Operation.Parameters == nil {
		o.Operation.Parameters = []*v3.Parameter{}
	}

	o.AddHandler(ps)

	for _, p := range ps.Parameters {
		o.Operation.Parameters = append(o.Operation.Parameters, p.Parameter)
	}

	return o
}

// Response adds a response to the operation.
func (o *Operation) Response(code string, cb func(r *Response)) *Operation {
	if o.Operation.Responses == nil {
		o.Operation.Responses = &v3.Responses{}
	}

	if o.Operation.Responses.Codes == nil {
		o.Operation.Responses.Codes = orderedmap.New[string, *v3.Response]()
	}

	codes := o.Operation.Responses.Codes
	if _, hasCode := codes.Get(code); !hasCode {
		codes.Set(code, &v3.Response{})
	}

	res := &Response{Response: codes.GetOrZero(code)}
	o.AddHandler(res)

	cb(res)

	return o
}

// SecurityRequirement configures the security scopes for this operation. The key in
// the map is the security scheme name and the value is the list of scopes.
func (o *Operation) SecurityRequirement(reqs map[string][]string) *Operation {
	if o.Operation.Security == nil {
		o.Operation.Security = []*base.SecurityRequirement{}
	}

	o.Operation.Security = append(o.Operation.Security, &base.SecurityRequirement{
		Requirements: orderedmap.ToOrderedMap(reqs),
	})

	return o
}
