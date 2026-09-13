package arrest

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"maps"
	"path"
	"reflect"
	"slices"
	"sort"
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
	makeRefs      map[string]*base.SchemaProxy
	componentRefs map[string]*base.SchemaProxy // refs that should be registered as components
	inProcess     map[reflect.Type]bool
}

// PolymorphicField represents a field that participates in polymorphic composition
type PolymorphicField struct {
	Field           reflect.StructField
	Type            reflect.Type
	CompositionType string // oneOf, anyOf, allOf
	Mapping         string // discriminator mapping alias
	RefName         string // component reference name if any
	IsInline        bool
}

// PolymorphicInfo contains information about a polymorphic struct
type PolymorphicInfo struct {
	DiscriminatorField reflect.StructField
	DefaultMapping     string
	CompositionType    string // oneOf, anyOf, allOf
	Fields             []PolymorphicField
}

//go:noinline
func newRefMapper(prefix string) *refMapper {
	return &refMapper{
		makeRefs:      make(map[string]*base.SchemaProxy),
		componentRefs: make(map[string]*base.SchemaProxy),
		inProcess:     make(map[reflect.Type]bool),
	}
}

// detectPolymorphicStruct examines a struct type to determine if it uses polymorphic tags
func detectPolymorphicStruct(t reflect.Type) (*PolymorphicInfo, bool) {
	var discriminatorField *reflect.StructField
	var defaultMapping string
	var compositionType string
	var polymorphFields []PolymorphicField

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue // skip unexported fields
		}

		info := NewTagInfo(field.Tag)

		// Check for discriminator field
		if info.IsDiscriminator() {
			if discriminatorField != nil {
				// Multiple discriminators not allowed
				return nil, false
			}
			discriminatorField = &field
			defaultMapping = info.GetDefaultMapping()
		}

		// Check for polymorphic composition fields
		fieldCompositionType := info.GetPolymorphType()
		if fieldCompositionType != "" {
			if compositionType == "" {
				compositionType = fieldCompositionType
			} else if compositionType != fieldCompositionType {
				// Mixed composition types not allowed
				return nil, false
			}

			fieldType := field.Type
			// Handle pointer types
			if fieldType.Kind() == reflect.Ptr {
				fieldType = fieldType.Elem()
			}

			polymorphFields = append(polymorphFields, PolymorphicField{
				Field:           field,
				Type:            fieldType,
				CompositionType: fieldCompositionType,
				Mapping:         info.GetMapping(),
				RefName:         info.RefName(),
				IsInline:        info.IsInline(),
			})
		}
	}

	// Must have both discriminator and polymorphic fields to be valid
	if discriminatorField == nil || len(polymorphFields) == 0 {
		return nil, false
	}

	return &PolymorphicInfo{
		DiscriminatorField: *discriminatorField,
		DefaultMapping:     defaultMapping,
		CompositionType:    compositionType,
		Fields:             polymorphFields,
	}, true
}

// buildPolymorphicSchema creates a polymorphic schema based on PolymorphicInfo
//
//go:noinline
func buildPolymorphicSchema(info *PolymorphicInfo, makeRefs *refMapper, skipDoc bool) (*base.SchemaProxy, error) {
	// Create schemas for each polymorphic field
	var schemas []*base.SchemaProxy
	mappingEntries := make(map[string]string)

	for _, field := range info.Fields {
		var fieldSchema *base.SchemaProxy
		var err error

		if field.IsInline || field.RefName != "" {
			// For inline fields or fields with refName, create the schema directly
			fieldSchema, err = makeSchemaProxy(field.Type, makeRefs, skipDoc)
			if err != nil {
				return nil, fmt.Errorf("failed to create schema for polymorphic field %s: %v", field.Field.Name, err)
			}
		} else {
			// For non-inline, non-ref fields, create a schema with the field as a property
			fieldProps := orderedmap.New[string, *base.SchemaProxy]()

			fSchema, err := makeSchemaProxy(field.Type, makeRefs, skipDoc)
			if err != nil {
				return nil, fmt.Errorf("failed to create schema for polymorphic field %s: %v", field.Field.Name, err)
			}

			fieldName := field.Field.Name
			fieldInfo := NewTagInfo(field.Field.Tag)
			if fieldInfo.HasName() {
				fieldName = fieldInfo.Name()
			}

			fieldProps.Set(fieldName, fSchema)
			var required []string
			if fieldInfo.IsRequired(field.Field.Type) {
				required = []string{fieldName}
			}
			fieldSchema = base.CreateSchemaProxy(&base.Schema{
				Type:       []string{"object"},
				Properties: fieldProps,
				Required:   required,
			})
		}

		// Handle component references
		if field.RefName != "" {
			ref := makeRefs.makeComponentRef(field.RefName, field.Type, fieldSchema)
			fieldSchema = base.CreateSchemaProxyRef(ref)
		}

		schemas = append(schemas, fieldSchema)

		// Add mapping entry if specified
		if field.Mapping != "" {
			var refTarget string
			if field.RefName != "" {
				// Use makeName to create a properly qualified type name that can be mapped
				typeName := makeName(field.RefName, field.Type, "")
				refTarget = "#/components/schemas/" + typeName
			} else {
				// For inline schemas, we'll need to create a component ref
				typeName := makeName("", field.Type, "")
				refTarget = "#/components/schemas/" + typeName
			}
			mappingEntries[field.Mapping] = refTarget
		}
	}

	// Create the polymorphic schema
	var schema *base.Schema
	switch info.CompositionType {
	case "oneOf":
		schema = &base.Schema{OneOf: schemas}
	case "anyOf":
		schema = &base.Schema{AnyOf: schemas}
	case "allOf":
		schema = &base.Schema{AllOf: schemas}
	default:
		return nil, fmt.Errorf("unsupported composition type: %s", info.CompositionType)
	}

	// Add discriminator if we have mappings or default mapping
	if len(mappingEntries) > 0 || info.DefaultMapping != "" {
		discriminatorFieldName := info.DiscriminatorField.Name
		discriminatorInfo := NewTagInfo(info.DiscriminatorField.Tag)
		if discriminatorInfo.HasName() {
			discriminatorFieldName = discriminatorInfo.Name()
		}

		mapping := orderedmap.New[string, string]()
		// Sort keys to ensure deterministic order
		aliases := make([]string, 0, len(mappingEntries))
		for alias := range mappingEntries {
			aliases = append(aliases, alias)
		}
		sort.Strings(aliases)

		for _, alias := range aliases {
			mapping.Set(alias, mappingEntries[alias])
		}

		schema.Discriminator = &base.Discriminator{
			PropertyName:   discriminatorFieldName,
			DefaultMapping: info.DefaultMapping,
			Mapping:        mapping,
		}
	}

	return base.CreateSchemaProxy(schema), nil
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

//go:noinline
func (m *refMapper) makeComponentRef(refName string, t reflect.Type, sp *base.SchemaProxy) string {
	name := makeName(refName, t, "")
	m.makeRefs[name] = sp
	m.componentRefs[name] = sp
	return "#/components/schemas/" + name
}

// Model provides DSL methods for creating OpenAPI schema objects based on Go
// types. The libopenapi schema underneath is available through OpenAPISchema.
type Model struct {
	// Name is the fully qualified Go type name the model stands for, or a
	// bare label like "OneOf" for a composed model, or "" for an anonymous
	// schema.
	Name string

	state any // *modelState; see state.go

	ErrHelper
}

// AnyOf associates a list of enumerations with the model.
//
//go:noinline
func (m *Model) AnyOf(enums ...Enumeration) *Model {
	schema := schemaOf(m).Schema()
	schema.AnyOf = make([]*base.SchemaProxy, len(enums))
	for i, enum := range enums {
		schema.AnyOf[i] = base.CreateSchemaProxy(&base.Schema{
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
//
//go:noinline
func (m *Model) OneOf(enums ...Enumeration) *Model {
	schema := schemaOf(m).Schema()
	schema.OneOf = make([]*base.SchemaProxy, len(enums))
	for i, enum := range enums {
		schema.OneOf[i] = base.CreateSchemaProxy(&base.Schema{
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

// sanitizeComponentName replaces slashes with periods to conform to OpenAPI 3.0
// schema component name requirements: [A-Za-z0-9_.-]
func sanitizeComponentName(name string) string {
	return strings.ReplaceAll(name, "/", ".")
}

func MappedName(typName string, pkgMap []PackageMap) string {
	if typName == "" {
		return ""
	}

	for _, item := range pkgMap {
		oasPkg, goPkg := item.OpenAPIName, item.GoName
		if trimName := strings.TrimPrefix(typName, goPkg); trimName != typName && trimName[0] == '.' {
			return sanitizeComponentName(oasPkg + trimName)
		}
	}

	return sanitizeComponentName(typName)
}

func (m *Model) MappedName(pkgMap []PackageMap) string {
	return MappedName(m.Name, pkgMap)
}

// Description sets the schema's description.
//
//go:noinline
func (m *Model) Description(description string) *Model {
	schemaOf(m).Schema().Description = description
	return m
}

// IsReference reports whether the model is a $ref to a component rather than
// a schema of its own.
//
//go:noinline
func (m *Model) IsReference() bool {
	sp := schemaOf(m)
	return sp != nil && sp.IsReference()
}

// Discriminator configures the discriminator for polymorphic schemas.
// It takes a property name used to discriminate between schemas, a default mapping,
// and optional alias-to-value mapping pairs.
// The mappings parameter should contain pairs of strings: alias1, value1, alias2, value2, etc.
//
//go:noinline
func (m *Model) Discriminator(propertyName, defaultMapping string, mappings ...string) *Model {
	if len(mappings)%2 != 0 {
		return withErr(m, errors.New("discriminator mappings must be provided in pairs (alias, value)"))
	}

	// Create the mapping from the variadic arguments
	mapping := orderedmap.New[string, string]()
	for i := 0; i < len(mappings); i += 2 {
		alias := mappings[i]
		value := mappings[i+1]
		mapping.Set(alias, value)
	}

	// Create the discriminator
	discriminator := &base.Discriminator{
		PropertyName:   propertyName,
		DefaultMapping: defaultMapping,
		Mapping:        mapping,
	}

	schemaOf(m).Schema().Discriminator = discriminator
	return m
}

// sortedRefs yields refs in name order. Registering components from a Go
// map's own iteration order would list them differently on every run.
//
//go:noinline
func sortedRefs(refs map[string]*base.SchemaProxy) iter.Seq2[string, *base.SchemaProxy] {
	return func(yield func(string, *base.SchemaProxy) bool) {
		for _, name := range slices.Sorted(maps.Keys(refs)) {
			if !yield(name, refs[name]) {
				return
			}
		}
	}
}

//go:noinline
func makeSchemaProxyStruct(t reflect.Type, makeRefs *refMapper, skipDoc bool) (*base.SchemaProxy, error) {
	// Check if this is a polymorphic struct first
	if polymorphInfo, isPolymorphic := detectPolymorphicStruct(t); isPolymorphic {
		return buildPolymorphicSchema(polymorphInfo, makeRefs, skipDoc)
	}

	doc := ""
	fieldDocs := map[string]string{}
	if !skipDoc {
		doc, fieldDocs, _ = GoDocForStruct(t)
	}

	fieldProps := orderedmap.New[string, *base.SchemaProxy]()
	// requiredByName tracks requiredness per JSON property name so that a
	// later field with the same name (e.g. an outer field shadowing a promoted
	// embedded field) replaces rather than duplicates the earlier entry.
	requiredByName := map[string]bool{}
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
				// very permissive
				return base.CreateSchemaProxy(&base.Schema{}), err
			}

			// Fields promoted from a nil embedded pointer are omitted entirely
			// by encoding/json, so nothing promoted through a pointer is required.
			anonRequired := map[string]bool{}
			if fType.Kind() != reflect.Ptr {
				for _, name := range anonSchema.Schema().Required {
					anonRequired[name] = true
				}
			}
			for k, v := range anonSchema.Schema().Properties.FromOldest() {
				fieldProps.Set(k, v)
				requiredByName[k] = anonRequired[k]
			}

			continue
		} else {
			var err error
			fSchema, err = makeSchemaProxy(fType, makeRefs, skipDoc)
			if err != nil {
				// very permissive
				return base.CreateSchemaProxy(&base.Schema{}),
					fmt.Errorf("failed to resolve field named %q with Go type %q: %v", f.Name, fType.String(), err)
			}

			if fDescription != "" {
				fSchema.Schema().Description = fDescription
			}

			if fType.Kind() == reflect.Slice || fType.Kind() == reflect.Array {
				if elemRefName := info.ElemRefName(); elemRefName != "" {
					fElemSchema, err := makeSchemaProxy(fType.Elem(), makeRefs, skipDoc)
					if err != nil {
						// very permissive
						return base.CreateSchemaProxy(&base.Schema{}),
							fmt.Errorf("failed to resolve field named %q with Go type %q: %v", f.Name, fType.String(), err)
					}

					elemRef := makeRefs.makeComponentRef(elemRefName, fType.Elem(), fElemSchema)
					itemSchema := base.CreateSchemaProxyRef(elemRef)
					fSchema = base.CreateSchemaProxy(&base.Schema{
						Type:  []string{"array"},
						Items: &base.DynamicValue[*base.SchemaProxy, bool]{N: 0, A: itemSchema},
					})
				}
			}

			if refName := info.RefName(); refName != "" {
				ref := makeRefs.makeComponentRef(refName, fType, fSchema)
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
		requiredByName[fName] = info.IsRequired(fType)
	}

	required := make([]string, 0, len(requiredByName))
	for name := range fieldProps.KeysFromOldest() {
		if requiredByName[name] {
			required = append(required, name)
		}
	}

	schema := &base.Schema{
		Description: doc,
		Type:        []string{"object"},
		Properties:  fieldProps,
	}
	if len(required) > 0 {
		schema.Required = required
	}

	return base.CreateSchemaProxy(schema), nil
}

//go:noinline
func makeSchemaProxySlice(t reflect.Type, makeRefs *refMapper, skipDoc bool) (*base.SchemaProxy, error) {
	sp, err := makeSchemaProxy(t.Elem(), makeRefs, skipDoc)
	if err != nil {
		return base.CreateSchemaProxy(&base.Schema{
			Type: []string{"array"},
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

//go:noinline
func makeSchemaProxyMap(t reflect.Type, makeRefs *refMapper, skipDoc bool) (*base.SchemaProxy, error) {
	if t.Key().Kind() != reflect.String {
		// technically illegal, so we'll just return it as type unspecified
		return base.CreateSchemaProxy(&base.Schema{
			Type: []string{"object"},
			AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{
				N: 1,
				B: true,
			},
		}), nil
	}

	sp, err := makeSchemaProxy(t.Elem(), makeRefs, skipDoc)
	if err != nil {
		// if the error is ignored, this will render permissive
		return base.CreateSchemaProxy(&base.Schema{
			Type: []string{"object"},
			AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{
				N: 1,
				B: true,
			},
		}), fmt.Errorf("failed to resolve inner type of map: %v", err)
	}

	// render as an object that points to a specific type
	schema := base.CreateSchemaProxy(&base.Schema{
		Type: []string{"object"},
		AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{
			N: 0,
			A: sp,
		},
	})

	return schema, nil
}

//go:noinline
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
		// very permissive
		schema = base.CreateSchemaProxy(&base.Schema{})
	default:
		// very permissive
		schema = base.CreateSchemaProxy(&base.Schema{})
		err = ErrUnsupportedModelType
	}

	// If this is a named struct type that could be recursive, register it in makeRefs
	if shouldRegister && schema != nil && err == nil {
		name := makeName("", t, "")
		makeRefs.makeRefs[name] = schema
	}

	return schema, err
}

// ModelOption configures model creation behavior.
type ModelOption func(*modelConfig)

type modelConfig struct {
	asComponent   bool
	componentName string
}

// AsComponent registers the model as a schema component in the document.
func AsComponent() ModelOption {
	return func(c *modelConfig) {
		c.asComponent = true
	}
}

// WithComponentName registers the model as a schema component with a custom name.
func WithComponentName(name string) ModelOption {
	return func(c *modelConfig) {
		c.asComponent = true
		c.componentName = name
	}
}

// ModelFromReflect creates a new Model from a reflect.Type with document context.
//
//go:noinline
func ModelFromReflect(t reflect.Type, doc *Document, opts ...ModelOption) *Model {
	mr := newRefMapper(t.PkgPath())
	sp, err := makeSchemaProxy(t, mr, SkipDocumentation)
	name := typeName(t)
	m := withErr(newModel(name, sp, mr.makeRefs, mr.componentRefs), err)
	if sp == nil {
		panic(fmt.Sprintf("failed to create SchemaProxy for type %s: got nil", name))
	} else if sp.Schema() == nil {
		panic(fmt.Sprintf("SchemaProxy for type %s returned nil Schema", name))
	}

	// Add to document handlers
	doc.AddHandler(m)

	// Apply package mapping to schema references
	if slices.Contains(sp.Schema().Type, "object") ||
		sp.Schema().OneOf != nil ||
		sp.Schema().AnyOf != nil ||
		sp.Schema().AllOf != nil {
		remapSchemaRefs(context.TODO(), sp, doc.PkgMap)
	}
	for pkgName, sp := range mr.makeRefs {
		if (sp == nil) || (sp.Schema() == nil) {
			doc.AddError(fmt.Errorf("failed while remapping schema refs for model %s, the package for %s has not be given a schema proxy definition; this may happen when refName and elemRefName are not consistent set", m.Name, pkgName))
			continue
		}

		if slices.Contains(sp.Schema().Type, "object") ||
			sp.Schema().OneOf != nil ||
			sp.Schema().AnyOf != nil ||
			sp.Schema().AllOf != nil {
			remapSchemaRefs(context.TODO(), sp, doc.PkgMap)
		}
	}

	// Register component references (from refName tags) automatically
	if len(mr.componentRefs) > 0 {
		c := schemasOf(doc)
		for goPkg, sp := range sortedRefs(mr.componentRefs) {
			if (sp == nil) || (sp.Schema() == nil) {
				doc.AddError(fmt.Errorf("failed while registering component reference for model %s, the package %s has not be given a schema proxy object; this may happen when refName and elemRefName are not consistently set", m.Name, goPkg))
				continue
			}

			componentFqn := MappedName(goPkg, doc.PkgMap)
			// Apply package mapping to component schemas
			if slices.Contains(sp.Schema().Type, "object") ||
				sp.Schema().OneOf != nil ||
				sp.Schema().AnyOf != nil ||
				sp.Schema().AllOf != nil {
				remapSchemaRefs(context.TODO(), sp, doc.PkgMap)
			}
			c.Set(componentFqn, sp)
		}
	}

	// Process options
	config := &modelConfig{}
	for _, opt := range opts {
		opt(config)
	}

	// Register as component if requested
	if config.asComponent {
		switch {
		case config.componentName != "" || isNamedType(t):
			doc.SchemaComponent(config.componentName, m)
		case isSliceOfNamed(t):
			// An unnamed slice has nothing to register under, but its element
			// does: register the element and make this an array of references.
			elemModel := ModelFromReflect(elemType(t), doc, AsComponent())
			stateOf(m).schema = base.CreateSchemaProxy(&base.Schema{
				Type: []string{"array"},
				Items: &base.DynamicValue[*base.SchemaProxy, bool]{
					A: schemaOf(SchemaRef(elemModel.MappedName(doc.PkgMap))),
				},
			})
		}
		// Any other unnamed type cannot be a component and stays inline.
	}

	return m
}

// derefType strips any pointers from t.
func derefType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

// typeName returns the fully qualified Go name of t (pointers stripped), as
// used for Model.Name. Unnamed and builtin types yield "".
func typeName(t reflect.Type) string {
	t = derefType(t)
	if !isNamedType(t) {
		return ""
	}
	return t.PkgPath() + "." + t.Name()
}

// isNamedType reports whether t (pointers stripped) is a type declared in
// some package, i.e. one that has a name to register a component under.
// Builtins like string have a name but no package and are excluded.
func isNamedType(t reflect.Type) bool {
	t = derefType(t)
	return t.PkgPath() != "" && t.Name() != ""
}

// elemType returns the element type of a slice or array with pointers
// stripped.
func elemType(t reflect.Type) reflect.Type {
	return derefType(t.Elem())
}

// isSliceOfNamed reports whether t is a slice or array whose element (after
// stripping pointers) is a named type.
func isSliceOfNamed(t reflect.Type) bool {
	if t.Kind() != reflect.Slice && t.Kind() != reflect.Array {
		return false
	}
	return isNamedType(elemType(t))
}

// ModelFrom creates a new Model from a type with document context.
func ModelFrom[T any](doc *Document, opts ...ModelOption) *Model {
	var t T
	return ModelFromReflect(reflect.TypeOf(t), doc, opts...)
}

// ModelFromReflectOnly creates a new Model from a reflect.Type without document context.
// This is intended for simple cases like parameters where document registration is not needed.
//
//go:noinline
func ModelFromReflectOnly(t reflect.Type) *Model {
	mr := newRefMapper(t.PkgPath())
	sp, err := makeSchemaProxy(t, mr, SkipDocumentation)
	name := typeName(t)
	m := withErr(newModel(name, sp, mr.makeRefs, mr.componentRefs), err)
	if sp == nil {
		panic(fmt.Sprintf("failed to create SchemaProxy for type %s: got nil", name))
	} else if sp.Schema() == nil {
		panic(fmt.Sprintf("SchemaProxy for type %s returned nil Schema", name))
	}
	return m
}

// ModelFromOnly creates a new Model from a type without document context.
// This is intended for simple cases like parameters where document registration is not needed.
func ModelFromOnly[T any]() *Model {
	var t T
	return ModelFromReflectOnly(reflect.TypeOf(t))
}

// SchemaRef creates a model that is a $ref to the schema component fqn.
//
//go:noinline
func SchemaRef(fqn string) *Model {
	sanitizedName := sanitizeComponentName(fqn)
	return newModel(fqn, base.CreateSchemaProxyRef("#"+path.Join("/components/schemas", sanitizedName)), nil, nil)
}

// OneOfTheseModels creates a model that represents a oneOf composition of the provided models.
// This is used for polymorphic schemas where exactly one of the provided schemas should match.
//
//go:noinline
func OneOfTheseModels(doc *Document, models ...*Model) *Model {
	if len(models) == 0 {
		return withErr(newModel("OneOf", base.CreateSchemaProxy(&base.Schema{}),
			make(map[string]*base.SchemaProxy), make(map[string]*base.SchemaProxy)), ErrUnsupportedModelType)
	}

	// Create SchemaProxy slice for OneOf
	oneOfSchemas := make([]*base.SchemaProxy, len(models))
	allMakeRefs := make(map[string]*base.SchemaProxy)
	allComponentRefs := make(map[string]*base.SchemaProxy)

	// Collect all errors from input models
	var firstErr error
	for i, model := range models {
		if model.Err() != nil && firstErr == nil {
			firstErr = model.Err()
		}

		st := stateOf(model)
		oneOfSchemas[i] = st.schema

		// Merge refs from all models
		for k, v := range st.makeRefs {
			allMakeRefs[k] = v
		}
		for k, v := range st.componentRefs {
			allComponentRefs[k] = v
		}
	}

	// Create the composed schema
	schema := &base.Schema{
		OneOf: oneOfSchemas,
	}

	m := withErr(newModel("OneOf", base.CreateSchemaProxy(schema), allMakeRefs, allComponentRefs), firstErr)

	// Add to document handlers
	doc.AddHandler(m)

	return m
}

// AnyOfTheseModels creates a model that represents an anyOf composition of the provided models.
// This is used for polymorphic schemas where any of the provided schemas can match.
//
//go:noinline
func AnyOfTheseModels(doc *Document, models ...*Model) *Model {
	if len(models) == 0 {
		return withErr(newModel("AnyOf", base.CreateSchemaProxy(&base.Schema{}),
			make(map[string]*base.SchemaProxy), make(map[string]*base.SchemaProxy)), ErrUnsupportedModelType)
	}

	// Create SchemaProxy slice for AnyOf
	anyOfSchemas := make([]*base.SchemaProxy, len(models))
	allMakeRefs := make(map[string]*base.SchemaProxy)
	allComponentRefs := make(map[string]*base.SchemaProxy)

	// Collect all errors from input models
	var firstErr error
	for i, model := range models {
		if model.Err() != nil && firstErr == nil {
			firstErr = model.Err()
		}

		st := stateOf(model)
		anyOfSchemas[i] = st.schema

		// Merge refs from all models
		for k, v := range st.makeRefs {
			allMakeRefs[k] = v
		}
		for k, v := range st.componentRefs {
			allComponentRefs[k] = v
		}
	}

	// Create the composed schema
	schema := &base.Schema{
		AnyOf: anyOfSchemas,
	}

	m := withErr(newModel("AnyOf", base.CreateSchemaProxy(schema), allMakeRefs, allComponentRefs), firstErr)

	// Add to document handlers
	doc.AddHandler(m)

	return m
}

// AllOfTheseModels creates a model that represents an allOf composition of the provided models.
// This is used for polymorphic schemas where all of the provided schemas must match.
//
//go:noinline
func AllOfTheseModels(doc *Document, models ...*Model) *Model {
	if len(models) == 0 {
		return withErr(newModel("AllOf", base.CreateSchemaProxy(&base.Schema{}),
			make(map[string]*base.SchemaProxy), make(map[string]*base.SchemaProxy)), ErrUnsupportedModelType)
	}

	// Create SchemaProxy slice for AllOf
	allOfSchemas := make([]*base.SchemaProxy, len(models))
	allMakeRefs := make(map[string]*base.SchemaProxy)
	allComponentRefs := make(map[string]*base.SchemaProxy)

	// Collect all errors from input models
	var firstErr error
	for i, model := range models {
		if model.Err() != nil && firstErr == nil {
			firstErr = model.Err()
		}

		st := stateOf(model)
		allOfSchemas[i] = st.schema

		// Merge refs from all models
		for k, v := range st.makeRefs {
			allMakeRefs[k] = v
		}
		for k, v := range st.componentRefs {
			allComponentRefs[k] = v
		}
	}

	// Create the composed schema
	schema := &base.Schema{
		AllOf: allOfSchemas,
	}

	m := withErr(newModel("AllOf", base.CreateSchemaProxy(schema), allMakeRefs, allComponentRefs), firstErr)

	// Add to document handlers
	doc.AddHandler(m)

	return m
}
