package arrest

import (
	"context"
	"errors"
	"reflect"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// ErrUnsupportedParameterType is returned when a parameter is created from an
// unsupported type.
var ErrUnsupportedParameterType = errors.New("unsupported parameter type")

// Parameter provides DSL methods for creating individual OpenAPI parameters.
type Parameter struct {
	Parameter *v3.Parameter

	ErrHelper
}

// Parameters provides DSL methods for creating multiple OpenAPI parameters.
type Parameters struct {
	Parameters []*Parameter

	ErrHelper
}

// ParameterFromReflect creates a new Parameter from a reflect.Type.
func ParameterFromReflect(t reflect.Type) *Parameter {
	p := &Parameter{
		Parameter: &v3.Parameter{},
	}

	m := ModelFromReflectOnly(t)

	p.AddHandler(m)
	p.Parameter.Schema = m.SchemaProxy
	return p
}

// ParameterFrom creates a new Parameter from a type.
func ParameterFrom[T any]() *Parameter {
	var t T
	return ParameterFromReflect(reflect.TypeOf(t))
}

// ParametersFromReflect creates a new Parameters from a reflect.Type. Given the
// reflect.Type for a function, it will use the function parameters to create
// the base list of parameters. You will need to use the P() method to access
// the parameters and set names in that case.
func ParametersFromReflect(t reflect.Type) *Parameters {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.Func:
		return parametersFromFunc(t)
	case reflect.Struct:
		return parametersFromStruct(t)
	default:
		return withErr(&Parameters{}, ErrUnsupportedParameterType)
	}
}

func parametersFromFunc(t reflect.Type) *Parameters {
	ps := &Parameters{
		Parameters: make([]*Parameter, 0, t.NumIn()),
	}

	for i := range t.NumIn() {
		// Ignore context variables
		if t.In(i).Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
			continue
		}

		p := ParameterFromReflect(t.In(i))

		ps.AddHandler(p)
		ps.Parameters = append(ps.Parameters, p)
	}

	return ps
}

func parametersFromStruct(t reflect.Type) *Parameters {
	_, fieldDocs, _ := GoDocForStruct(t)

	ps := &Parameters{
		Parameters: make([]*Parameter, 0, t.NumField()),
	}

	for i := range t.NumField() {
		f := t.Field(i)

		info := NewTagInfo(f.Tag)
		if info.IsIgnored() {
			continue
		}

		fIn := "query"
		if info.HasIn() {
			fIn = info.In()
		}

		fName := f.Name
		if info.HasName() {
			fName = info.Name()
		}

		fDescription := ""
		if fieldDocs != nil {
			fDescription = fieldDocs[fName]
		}

		fReplaceType := info.ReplacementType()

		var p *Parameter
		if fReplaceType != "" {
			// Create parameter with custom type override
			p = &Parameter{
				Parameter: &v3.Parameter{},
			}

			// Create schema with the replacement type
			p.Parameter.Schema = base.CreateSchemaProxy(&base.Schema{
				Description: fDescription,
				Type:        []string{fReplaceType},
			})
		} else {
			// Use default reflection-based parameter creation
			p = ParameterFromReflect(f.Type)
		}

		p = p.Name(fName).
			In(fIn).
			Description(fDescription)

		if fIn == "path" {
			p = p.Required()
		}

		ps.AddHandler(p)
		ps.Parameters = append(ps.Parameters, p)
	}

	return ps
}

// ParametersFrom creates a new Parameters from a type.
func ParametersFrom[T any]() *Parameters {
	var t T
	return ParametersFromReflect(reflect.TypeOf(t))
}

// NParameters creates a new Parameters with the given number of parameters.
func NParameters(n int) *Parameters {
	ps := &Parameters{
		Parameters: make([]*Parameter, n),
	}

	for i := range ps.Parameters {
		ps.Parameters[i] = &Parameter{Parameter: &v3.Parameter{}}
	}

	return ps
}

// P returns the parameter at the given index and calls the callback with it.
func (p *Parameters) P(idx int, cb func(p *Parameter)) *Parameters {
	cb(p.Parameters[idx])
	return p
}

// Name sets the name of the parameter.
func (p *Parameter) Name(name string) *Parameter {
	p.Parameter.Name = name
	return p
}

// In sets the location of the parameter. Use this method when the In is not
// one of the usual values like "query", "path", "header", or "cookie". In those
// cases, you should prefer the InQuery(), InPath(), InHeader(), or InCookie()
// methods instead.
func (p *Parameter) In(in string) *Parameter {
	p.Parameter.In = in
	return p
}

// InQuery sets the location of the parameter to "query".
func (p *Parameter) InQuery() *Parameter {
	p.Parameter.In = "query"
	return p
}

// InPath sets the location of the parameter to "path".
func (p *Parameter) InPath() *Parameter {
	p.Parameter.In = "path"
	return p
}

// InHeader sets the location of the parameter to "header".
func (p *Parameter) InHeader() *Parameter {
	p.Parameter.In = "header"
	return p
}

// InCookie sets the location of the parameter to "cookie".
func (p *Parameter) InCookie() *Parameter {
	p.Parameter.In = "cookie"
	return p
}

// Required marks the parameter as required.
func (p *Parameter) Required() *Parameter {
	req := true
	p.Parameter.Required = &req
	return p
}

// Description sets the description of the parameter.
func (p *Parameter) Description(description string) *Parameter {
	p.Parameter.Description = description
	return p
}

// Model sets the schema of the parameter.
func (p *Parameter) Model(m *Model) *Parameter {
	p.AddHandler(m)
	p.Parameter.Schema = m.SchemaProxy
	return p
}
