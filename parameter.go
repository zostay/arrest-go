package arrest

import (
	"errors"
	"reflect"

	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

var ErrUnsupportedParameterType = errors.New("unsupported parameter type")

type Parameter struct {
	Parameter *v3.Parameter

	ErrHelper
}

type Parameters struct {
	Parameters []*Parameter

	ErrHelper
}

func ParameterFromReflect(t reflect.Type) *Parameter {
	p := &Parameter{
		Parameter: &v3.Parameter{},
	}

	m := ModelFromReflect(t)

	p.AddHandler(m)
	p.Parameter.Schema = m.SchemaProxy
	return p
}

func ParameterFrom[T any]() *Parameter {
	var t T
	return ParameterFromReflect(reflect.TypeOf(t))
}

func ParametersFromReflect(t reflect.Type) *Parameters {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Func {
		return withErr(&Parameters{}, ErrUnsupportedParameterType)
	}

	ps := &Parameters{
		Parameters: make([]*Parameter, t.NumIn()),
	}

	for i := range t.NumIn() {
		p := ParameterFromReflect(t.In(i))

		ps.AddHandler(p)
		ps.Parameters[i] = p
	}

	return ps
}

func ParametersFrom[T any]() *Parameters {
	var t T
	return ParametersFromReflect(reflect.TypeOf(t))
}

func (p *Parameters) P(idx int, cb func(p *Parameter)) *Parameters {
	cb(p.Parameters[idx])
	return p
}

func (p *Parameter) Name(name string) *Parameter {
	p.Parameter.Name = name
	return p
}

func (p *Parameter) In(in string) *Parameter {
	p.Parameter.In = in
	return p
}

func (p *Parameter) Required() *Parameter {
	req := true
	p.Parameter.Required = &req
	return p
}

func (p *Parameter) Description(description string) *Parameter {
	p.Parameter.Description = description
	return p
}
