package arrest

import (
	"errors"
	"fmt"
	"path"
	"reflect"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/orderedmap"
)

// ErrUnsupportedModelType is returned when the model type is not supported.
var ErrUnsupportedModelType = errors.New("unsupported model type")

// Model provides DSL methods for creating OpenAPI schema objects based on Go
// types.
type Model struct {
	SchemaProxy *base.SchemaProxy

	ErrHelper
}

func makeSchemaProxyStruct(t reflect.Type) (*base.SchemaProxy, error) {
	fieldProps := orderedmap.New[string, *base.SchemaProxy]()
	for i := range t.NumField() {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue
		}

		fName := f.Name
		fType := f.Type
		arrestTag := f.Tag.Get("arrest")
		jsonTag := f.Tag.Get("json")

		if arrestTag != "" {
			fName = arrestTag
		} else if jsonTag != "" {
			fName = jsonTag
		}

		fSchema, err := makeSchemaProxy(fType)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve field named %q: %v", f.Name, err)
		}

		fieldProps.Set(fName, fSchema)
	}

	schema := &base.Schema{
		Type:       []string{"object"},
		Properties: fieldProps,
	}

	return base.CreateSchemaProxy(schema), nil
}

func makeSchemaProxySlice(t reflect.Type) (*base.SchemaProxy, error) {
	sp, err := makeSchemaProxy(t.Elem())
	if err != nil {
		return nil, fmt.Errorf("failed to resolve inner type of array or slice: %v", err)
	}

	schema := base.CreateSchemaProxy(&base.Schema{
		Type:  []string{"array"},
		Items: &base.DynamicValue[*base.SchemaProxy, bool]{N: 0, A: sp},
	})

	if t.Kind() == reflect.Array {
		maxLen := int64(t.Len())
		schema.Schema().MaxItems = &maxLen
	}

	return schema, nil
}

func makeSchemaProxy(t reflect.Type) (*base.SchemaProxy, error) {
	switch t.Kind() {
	case reflect.Struct:
		return makeSchemaProxyStruct(t)
	case reflect.Slice, reflect.Array:
		return makeSchemaProxySlice(t)
	case reflect.Ptr:
		return makeSchemaProxy(t.Elem())
	case reflect.Bool:
		return base.CreateSchemaProxy(&base.Schema{
			Type: []string{"boolean"},
		}), nil
	case reflect.String:
		return base.CreateSchemaProxy(&base.Schema{
			Type: []string{"string"},
		}), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		return base.CreateSchemaProxy(&base.Schema{
			Type:   []string{"integer"},
			Format: "int32",
		}), nil
	case reflect.Int64:
		return base.CreateSchemaProxy(&base.Schema{
			Type:   []string{"integer"},
			Format: "int64",
		}), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return base.CreateSchemaProxy(&base.Schema{
			Type: []string{"integer"},
		}), nil
	case reflect.Float32:
		return base.CreateSchemaProxy(&base.Schema{
			Type:   []string{"number"},
			Format: "float",
		}), nil
	case reflect.Float64:
		return base.CreateSchemaProxy(&base.Schema{
			Type:   []string{"number"},
			Format: "double",
		}), nil
	default:
		return nil, ErrUnsupportedModelType
	}
}

// ModelFromReflect creates a new Model from a reflect.Type.
func ModelFromReflect(t reflect.Type) *Model {
	sp, err := makeSchemaProxy(t)
	return withErr(&Model{SchemaProxy: sp}, err)
}

// ModelFrom creates a new Model from a type.
func ModelFrom[T any]() *Model {
	var t T
	return ModelFromReflect(reflect.TypeOf(t))
}

func SchemaRef(fqn string) *Model {
	return &Model{
		SchemaProxy: base.CreateSchemaProxyRef("#" + path.Join("/components/schemas", fqn)),
	}
}
