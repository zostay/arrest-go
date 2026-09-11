package arrest_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zostay/arrest-go"
	"github.com/zostay/arrest-go/internal/test"
)

func TestGoDocForStruct(t *testing.T) {
	t.Parallel()

	doc, flds, err := arrest.GoDocForStruct(reflect.TypeOf(test.DocStruct{}))
	require.NoError(t, err)

	assert.Equal(t, "DocStruct is a structure with documentation.", doc)

	assert.Equal(t, map[string]string{
		"Foo":      "Foo is a field.",
		"Bar":      "Bar is also a field.",
		"tagged":   "tagged is renamed by its json tag.",
		"renamed":  "renamed is renamed by its openapi tag.",
		"Trailing": "Trailing is documented on the same line.",
		"Both1":    "Both share one comment.",
		"Both2":    "Both share one comment.",
	}, flds)
}

func TestGoDocForStruct_SadNotAStruct(t *testing.T) {
	t.Parallel()

	_, _, err := arrest.GoDocForStruct(reflect.TypeOf(1))
	assert.ErrorContains(t, err, "expected a struct type")
}

func BenchmarkGoDocForStruct(b *testing.B) {
	b.ReportAllocs()
	typeToTest := reflect.TypeOf(test.DocStruct{})
	for i := 0; i < b.N; i++ {
		_, _, err := arrest.GoDocForStruct(typeToTest)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

const expected_FieldDocs = `openapi: 3.1.0
info:
  title: test
paths:
  /items/{id}:
    put:
      parameters:
        - name: since
          in: query
          description: |-
            since is an RFC 3339 timestamp. When given, the response includes
            deleted items.
          schema:
            type: string
        - name: id
          in: path
          required: true
          description: id identifies the item.
          schema:
            type: string
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: object
                description: DocStruct is a structure with documentation.
                properties:
                  Foo:
                    type: string
                    description: Foo is a field.
                  Bar:
                    type: object
                    description: Bar is also a field.
                    properties:
                      Baz:
                        type: string
                        description: Baz is a field within a field.
                    required:
                      - Baz
                  tagged:
                    type: string
                    description: tagged is renamed by its json tag.
                  renamed:
                    type: string
                    description: renamed is renamed by its openapi tag.
                  Trailing:
                    type: string
                    description: Trailing is documented on the same line.
                  Both1:
                    type: string
                    description: Both share one comment.
                  Both2:
                    type: string
                    description: Both share one comment.
                required:
                  - Foo
                  - tagged
                  - renamed
                  - Trailing
                  - Both1
                  - Both2
`

// Field-level doc comments reach schema properties and parameters.
func TestGoDoc_FieldDescriptions(t *testing.T) {
	t.Parallel()

	doc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	doc.Put("/items/{id}").
		Parameters(arrest.ParametersFromReflect(reflect.TypeOf(test.DocRequest{}))).
		Response("200", func(r *arrest.Response) {
			r.Description("OK").Content("application/json", arrest.ModelFrom[test.DocStruct](doc))
		})

	require.NoError(t, doc.Err())
	oas, err := doc.OpenAPI.Render()
	require.NoError(t, err)
	assert.YAMLEq(t, expected_FieldDocs, string(oas))
}
