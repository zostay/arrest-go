package arrest

// SchemaComponent represents a registered schema component in a document.
// This is a result type used for querying existing components.
type SchemaComponent struct {
	schema *Model
	ref    *Model
	name   string
}

// NewSchemaComponent creates a new SchemaComponent result.
func NewSchemaComponent(name string, schema *Model, ref *Model) *SchemaComponent {
	return &SchemaComponent{
		name:   name,
		schema: schema,
		ref:    ref,
	}
}

// Name returns the component name as registered in the document.
func (s *SchemaComponent) Name() string {
	return s.name
}

// Schema returns the actual schema model.
func (s *SchemaComponent) Schema() *Model {
	return s.schema
}

// Ref returns the reference model for this component.
func (s *SchemaComponent) Ref() *Model {
	return s.ref
}
