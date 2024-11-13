package arrest

import (
	"errors"
	"fmt"
	"path"
	"reflect"
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/orderedmap"
)

// ErrUnsupportedModelType is returned when the model type is not supported.
var ErrUnsupportedModelType = errors.New("unsupported model type")

type refMapper struct {
	prefix   string
	makeRefs map[string]*base.SchemaProxy
}

func newRefMapper(prefix string) *refMapper {
	return &refMapper{
		prefix:   prefix,
		makeRefs: make(map[string]*base.SchemaProxy),
	}
}

func (m *refMapper) makeRef(fName string, sp *base.SchemaProxy) {
	m.makeRefs[m.prefix+"."+fName] = sp
}

// Model provides DSL methods for creating OpenAPI schema objects based on Go
// types.
type Model struct {
	Name        string
	SchemaProxy *base.SchemaProxy

	makeRefs map[string]*base.SchemaProxy

	ErrHelper
}

func MappedName(typName string, pkgMap map[string]string) string {
	if typName == "" {
		return ""
	}

	for oasPkg, goPkg := range pkgMap {
		if trimName := strings.TrimPrefix(typName, goPkg); trimName != typName && trimName[0] == '.' {
			return oasPkg + trimName
		}
	}

	return typName
}

func (m *Model) MappedName(pkgMap map[string]string) string {
	return MappedName(m.Name, pkgMap)
}

func (m *Model) Description(description string) *Model {
	m.SchemaProxy.Schema().Description = description
	return m
}

func (m *Model) ExtractChildRefs() map[string]*base.SchemaProxy {
	return m.makeRefs
}

func makeSchemaProxyStruct(t reflect.Type, makeRefs *refMapper) (*base.SchemaProxy, error) {
	doc, fieldDocs, _ := GoDocForStruct(t)

	fieldProps := orderedmap.New[string, *base.SchemaProxy]()
	for i := range t.NumField() {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue
		}

		fName := f.Name
		fType := f.Type

		info := NewTagInfo(f.Tag)
		if info.IsIgnored() {
			continue
		}

		if info.HasName() {
			fName = info.Name()
		}

		fDescription := ""
		if fieldDocs != nil {
			fDescription = fieldDocs[f.Name]
		}

		fReplaceType := info.ReplacementType()

		var fSchema *base.SchemaProxy
		if fReplaceType != "" {
			fSchema = base.CreateSchemaProxy(&base.Schema{
				Description: fDescription,
				Type:        []string{fReplaceType},
			})
		} else {
			var err error
			fSchema, err = makeSchemaProxy(fType, makeRefs)
			if err != nil {
				return base.CreateSchemaProxy(&base.Schema{
					Type: []string{"any"},
				}), fmt.Errorf("failed to resolve field named %q with Go type %q: %v", f.Name, fType.String(), err)
			}

			if fDescription != "" {
				fSchema.Schema().Description = fDescription
			}
		}

		if refName := info.RefName(); refName != "" {
			makeRefs.makeRef(refName, fSchema)
		}

		// TODO This would be super cool to implement.
		//schemaLow := fSchema.GoLow().Schema()
		//for key, value := range info.Props() {
		//	switch key {
		//	case "content-type":
		//		schemaLow.ContentMediaType = low.NodeReference[string]{Value: value}
		//	case "content-encoding":
		//		schemaLow.ContentEncoding = low.NodeReference[string]{Value: value}
		//	}
		//}

		fieldProps.Set(fName, fSchema)
	}

	schema := &base.Schema{
		Description: doc,
		Type:        []string{"object"},
		Properties:  fieldProps,
	}

	return base.CreateSchemaProxy(schema), nil
}

func makeSchemaProxySlice(t reflect.Type, makeRefs *refMapper) (*base.SchemaProxy, error) {
	sp, err := makeSchemaProxy(t.Elem(), makeRefs)
	if err != nil {
		return base.CreateSchemaProxy(&base.Schema{
			Type: []string{"any"},
		}), fmt.Errorf("failed to resolve inner type of array or slice: %v", err)
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

func makeSchemaProxyMap(t reflect.Type, makeRefs *refMapper) (*base.SchemaProxy, error) {
	sp, err := makeSchemaProxy(t.Elem(), makeRefs)
	if err != nil {
		return base.CreateSchemaProxy(&base.Schema{
			Type: []string{"any"},
		}), fmt.Errorf("failed to resolve inner type of map: %v", err)
	}

	schema := base.CreateSchemaProxy(&base.Schema{
		Type: []string{"object"},
		AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{
			N: 0,
			A: sp,
		},
	})

	return schema, nil
}

func makeSchemaProxy(t reflect.Type, makeRefs *refMapper) (*base.SchemaProxy, error) {
	switch t.Kind() {
	case reflect.Struct:
		return makeSchemaProxyStruct(t, makeRefs)
	case reflect.Slice, reflect.Array:
		return makeSchemaProxySlice(t, makeRefs)
	case reflect.Map:
		return makeSchemaProxyMap(t, makeRefs)
	case reflect.Ptr:
		return makeSchemaProxy(t.Elem(), makeRefs)
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
		return base.CreateSchemaProxy(&base.Schema{
			Type: []string{"any"},
		}), ErrUnsupportedModelType
	}
}

// ModelFromReflect creates a new Model from a reflect.Type.
func ModelFromReflect(t reflect.Type) *Model {
	mr := newRefMapper(t.PkgPath())
	sp, err := makeSchemaProxy(t, mr)
	name := strings.Join([]string{t.PkgPath(), t.Name()}, ".")
	m := withErr(&Model{Name: name, SchemaProxy: sp, makeRefs: mr.makeRefs}, err)
	if m.SchemaProxy == nil {
		panic("nope")
	} else if m.SchemaProxy.Schema() == nil {
		panic("noper")
	}
	return m
}

// ModelFrom creates a new Model from a type.
func ModelFrom[T any]() *Model {
	var t T
	return ModelFromReflect(reflect.TypeOf(t))
}

func SchemaRef(fqn string) *Model {
	return &Model{
		Name:        fqn,
		SchemaProxy: base.CreateSchemaProxyRef("#" + path.Join("/components/schemas", fqn)),
	}
}
