package arrest

import (
	"fmt"

	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

type Operation struct {
	Operation *v3.Operation

	ErrHelper
}

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

func (o *Operation) ResponseBody(code, mt string, model *Model) *Operation {
	if model.SchemaProxy == nil {
		return withErr(o, fmt.Errorf("model must be initialized"))
	}

	o.AddHandler(model)

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

	res := codes.GetOrZero(code)
	if res.Content == nil {
		res.Content = orderedmap.New[string, *v3.MediaType]()
	}

	res.Content.Set(mt, &v3.MediaType{Schema: model.SchemaProxy})

	return o
}

func (o *Operation) Summary(summary string) *Operation {
	o.Operation.Summary = summary
	return o
}

func (o *Operation) OperationID(id string) *Operation {
	o.Operation.OperationId = id
	return o
}

func (o *Operation) Tags(tags ...string) *Operation {
	o.Operation.Tags = append(o.Operation.Tags, tags...)
	return o
}

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
