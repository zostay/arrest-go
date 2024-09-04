package arrest

type SchemaComponent struct {
	schema *Model
	ref    *Model
}

func NewSchemaComponent(schema *Model, ref *Model) *SchemaComponent {
	return &SchemaComponent{
		schema: schema,
		ref:    ref,
	}
}

func (s *SchemaComponent) Description(description string) *SchemaComponent {
	s.schema.Description(description)
	return s
}

func (s *SchemaComponent) Schema() *Model {
	return s.schema
}

func (s *SchemaComponent) Ref() *Model {
	return s.ref
}
