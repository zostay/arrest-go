package arrest

import (
	"errors"
	"fmt"
	"path"
	"reflect"
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/orderedmap"
	"go.yaml.in/yaml/v4"
)

// SkipDocumentation is a global that can be set to true to skip generating
// documentation for models. This is useful during runtime as it greatly speeds
// up parsing and generation or to avoid using Go-based documentation in
// generated OpenAPI specs, but if OpenAPI documents are generated, they will be
// lacking documentation.
var SkipDocumentation = false

// ErrUnsupportedModelType is returned when the model type is not supported.
var ErrUnsupportedModelType = errors.New("unsupported model type")

type Enumeration struct {
	Const       any    `yaml:"const"`
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
}

type refMapper struct {
	makeRefs  map[string]*base.SchemaProxy
	inProcess map[reflect.Type]bool
}

func newRefMapper(prefix string) *refMapper {
	return &refMapper{
		makeRefs:  make(map[string]*base.SchemaProxy),
		inProcess: make(map[reflect.Type]bool),
	}
}

func makeName(refName string, t reflect.Type, defaultSuffix string) string {
	switch t.Kind() {
	case reflect.Ptr:
		return makeName(refName, t.Elem(), defaultSuffix)
	case reflect.Slice:
		return makeName(refName, t.Elem(), "List")
	default:
		if refName == "" {
			refName = t.Name() + defaultSuffix
		}
		return strings.Join([]string{t.PkgPath(), refName}, ".")
	}
}

func (m *refMapper) makeRef(refName string, t reflect.Type, sp *base.SchemaProxy) string {
	name := makeName(refName, t, "")
	m.makeRefs[name] = sp
	return "#/components/schemas/" + name
}

// Model provides DSL methods for creating OpenAPI schema objects based on Go
// types.
type Model struct {
	Name        string
	SchemaProxy *base.SchemaProxy

	makeRefs map[string]*base.SchemaProxy

	ErrHelper
}

// AnyOf associates a list of enumerations with the model.
func (m *Model) AnyOf(enums ...Enumeration) *Model {
	m.SchemaProxy.Schema().AnyOf = make([]*base.SchemaProxy, len(enums))
	for i, enum := range enums {
		m.SchemaProxy.Schema().AnyOf[i] = base.CreateSchemaProxy(&base.Schema{
			Title: enum.Title,
			Const: &yaml.Node{
				Kind:  yaml.ScalarNode,
				Value: fmt.Sprintf("%v", enum.Const),
			},
			Description: enum.Description,
		})
	}
	return m
}

// OneOf associates a list of enumerations with the model.
func (m *Model) OneOf(enums ...Enumeration) *Model {
	m.SchemaProxy.Schema().OneOf = make([]*base.SchemaProxy, len(enums))
	for i, enum := range enums {
		m.SchemaProxy.Schema().OneOf[i] = base.CreateSchemaProxy(&base.Schema{
			Title: enum.Title,
			Const: &yaml.Node{
				Kind:  yaml.ScalarNode,
				Value: fmt.Sprintf("%v", enum.Const),
			},
			Description: enum.Description,
		})
	}
	return m
}

func MappedName(typName string, pkgMap []PackageMap) string {
	if typName == "" {
		return ""
	}

	for _, item := range pkgMap {
		oasPkg, goPkg := item.OpenAPIName, item.GoName
		if trimName := strings.TrimPrefix(typName, goPkg); trimName != typName && trimName[0] == '.' {
			return oasPkg + trimName
		}
	}

	return typName
}

func (m *Model) MappedName(pkgMap []PackageMap) string {
	return MappedName(m.Name, pkgMap)
}

func (m *Model) Description(description string) *Model {
	m.SchemaProxy.Schema().Description = description
	return m
}

func (m *Model) ExtractChildRefs() map[string]*base.SchemaProxy {
	return m.makeRefs
}

func makeSchemaProxyStruct(t reflect.Type, makeRefs *refMapper, skipDoc bool) (*base.SchemaProxy, error) {
	doc := ""
	fieldDocs := map[string]string{}
	if !skipDoc {
		doc, fieldDocs, _ = GoDocForStruct(t)
	}

	fieldProps := orderedmap.New[string, *base.SchemaProxy]()
	for i := range t.NumField() {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue
		}

		fName := f.Name
		fType := f.Type

		info := NewTagInfo(f.Tag)
		if info.IsIgnored() || info.HasIn() {
			// either they are ignored or they are parameters that belong to the
			// path, query, headers, etc. (not here)
			continue
		}

		if info.HasName() {
			fName = info.Name()
		}

		fDescription := ""
		if fieldDocs != nil {
			fDescription = fieldDocs[fName]
		}

		fReplaceType := info.ReplacementType()

		var fSchema *base.SchemaProxy
		if fReplaceType != "" {
			fSchema = base.CreateSchemaProxy(&base.Schema{
				Description: fDescription,
				Type:        []string{fReplaceType},
			})
		} else if f.Anonymous {
			anonSchema, err := makeSchemaProxy(fType, makeRefs, skipDoc)
			if err != nil {
				return base.CreateSchemaProxy(&base.Schema{
					Type: []string{"any"},
				}), err
			}

			for k, v := range anonSchema.Schema().Properties.FromOldest() {
				fieldProps.Set(k, v)
			}

			continue
		} else {
			var err error
			fSchema, err = makeSchemaProxy(fType, makeRefs, skipDoc)
			if err != nil {
				return base.CreateSchemaProxy(&base.Schema{
					Type: []string{"any"},
				}), fmt.Errorf("failed to resolve field named %q with Go type %q: %v", f.Name, fType.String(), err)
			}

			if fDescription != "" {
				fSchema.Schema().Description = fDescription
			}

			if fType.Kind() == reflect.Slice || fType.Kind() == reflect.Array {
				if elemRefName := info.ElemRefName(); elemRefName != "" {
					fElemSchema, err := makeSchemaProxy(fType.Elem(), makeRefs, skipDoc)
					if err != nil {
						return base.CreateSchemaProxy(&base.Schema{
							Type: []string{"any"},
						}), fmt.Errorf("failed to resolve field named %q with Go type %q: %v", f.Name, fType.String(), err)
					}

					elemRef := makeRefs.makeRef(elemRefName, fType.Elem(), fElemSchema)
					itemSchema := base.CreateSchemaProxyRef(elemRef)
					fSchema = base.CreateSchemaProxy(&base.Schema{
						Type:  []string{"array"},
						Items: &base.DynamicValue[*base.SchemaProxy, bool]{N: 0, A: itemSchema},
					})
				}
			}

			if refName := info.RefName(); refName != "" {
				ref := makeRefs.makeRef(refName, fType, fSchema)
				fSchema = base.CreateSchemaProxyRef(ref)
			}
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

func makeSchemaProxySlice(t reflect.Type, makeRefs *refMapper, skipDoc bool) (*base.SchemaProxy, error) {
	sp, err := makeSchemaProxy(t.Elem(), makeRefs, skipDoc)
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

func makeSchemaProxyMap(t reflect.Type, makeRefs *refMapper, skipDoc bool) (*base.SchemaProxy, error) {
	sp, err := makeSchemaProxy(t.Elem(), makeRefs, skipDoc)
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

func makeSchemaProxy(t reflect.Type, makeRefs *refMapper, skipDoc bool) (*base.SchemaProxy, error) {
	// Check if this type is currently being processed to prevent infinite recursion
	if makeRefs.inProcess[t] {
		// Create a reference for this recursive type
		name := makeName("", t, "")
		ref := "#/components/schemas/" + name
		return base.CreateSchemaProxyRef(ref), nil
	}

	// Mark this type as being processed
	makeRefs.inProcess[t] = true
	defer func() {
		delete(makeRefs.inProcess, t)
	}()

	// For struct types that might be recursive, we need to pre-register them
	// This ensures that if we encounter a recursive reference, the schema will be available
	var shouldRegister bool
	actualType := t
	for actualType.Kind() == reflect.Ptr {
		actualType = actualType.Elem()
	}
	if actualType.Kind() == reflect.Struct && actualType.Name() != "" && actualType.Name() != "Time" {
		shouldRegister = true
	}

	// Create the schema
	var schema *base.SchemaProxy
	var err error

	switch t.Kind() {
	case reflect.Struct:
		if t.Name() == "Time" && t.PkgPath() == "time" {
			schema = base.CreateSchemaProxy(&base.Schema{
				Type:   []string{"string"},
				Format: "date-time",
			})
		} else {
			schema, err = makeSchemaProxyStruct(t, makeRefs, skipDoc)
		}
	case reflect.Slice, reflect.Array:
		schema, err = makeSchemaProxySlice(t, makeRefs, skipDoc)
	case reflect.Map:
		schema, err = makeSchemaProxyMap(t, makeRefs, skipDoc)
	case reflect.Ptr:
		schema, err = makeSchemaProxy(t.Elem(), makeRefs, skipDoc)
	case reflect.Bool:
		schema = base.CreateSchemaProxy(&base.Schema{
			Type: []string{"boolean"},
		})
	case reflect.String:
		schema = base.CreateSchemaProxy(&base.Schema{
			Type: []string{"string"},
		})
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		schema = base.CreateSchemaProxy(&base.Schema{
			Type:   []string{"integer"},
			Format: "int32",
		})
	case reflect.Int64:
		schema = base.CreateSchemaProxy(&base.Schema{
			Type:   []string{"integer"},
			Format: "int64",
		})
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		schema = base.CreateSchemaProxy(&base.Schema{
			Type: []string{"integer"},
		})
	case reflect.Float32:
		schema = base.CreateSchemaProxy(&base.Schema{
			Type:   []string{"number"},
			Format: "float",
		})
	case reflect.Float64:
		schema = base.CreateSchemaProxy(&base.Schema{
			Type:   []string{"number"},
			Format: "double",
		})
	case reflect.Interface:
		schema = base.CreateSchemaProxy(&base.Schema{
			Type: []string{"any"},
		})
	default:
		schema = base.CreateSchemaProxy(&base.Schema{
			Type: []string{"any"},
		})
		err = ErrUnsupportedModelType
	}

	// If this is a named struct type that could be recursive, register it in makeRefs
	if shouldRegister && schema != nil && err == nil {
		name := makeName("", t, "")
		makeRefs.makeRefs[name] = schema
	}

	return schema, err
}

// ModelFromReflect creates a new Model from a reflect.Type.
func ModelFromReflect(t reflect.Type) *Model {
	mr := newRefMapper(t.PkgPath())
	sp, err := makeSchemaProxy(t, mr, SkipDocumentation)
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
