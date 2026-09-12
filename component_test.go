package arrest_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zostay/arrest-go"
)

type TestReq struct {
	Test     TestType    `json:"test" openapi:",refName=TestType"`
	AlsoTest []Test2Type `json:"alsoTest" openapi:",elemRefName=Test2Type"`
}

type TestType struct {
	Field string `json:"field"`
}

type Test2Type struct {
	Field string `json:"field"`
}

func TestComponentRefFromTagAlone(t *testing.T) {
	t.Parallel()

	const expected = `openapi: 3.1.0
info:
  title: ComponentRefFromTagAlone
paths:
  /test:
    post:
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                test:
                  $ref: '#/components/schemas/test.v1.TestType'
                alsoTest:
                  type: array
                  items:
                    $ref: '#/components/schemas/test.v1.Test2Type'
              required:
                - test
                - alsoTest
components:
  schemas:
    test.v1.TestType:
      type: object
      properties:
        field:
          type: string
      required:
        - field
    test.v1.Test2Type:
      type: object
      properties:
        field:
          type: string
      required:
        - field
`

	doc, err := arrest.NewDocument("ComponentRefFromTagAlone")
	if err != nil {
		t.Fatalf("could not create document: %v", err)
	}

	doc.PackageMap(
		"test.v1", "github.com/zostay/arrest-go",
		"test.v1", "github.com/zostay/arrest-go_test",
		"test.v1", "command-line-arguments_test",
	)

	doc.Post("/test").
		RequestBody("application/json", arrest.ModelFrom[TestReq](doc))

	assert.NoError(t, doc.Err())
	got, err := doc.Render()
	assert.NoError(t, err)
	assert.YAMLEq(t, expected, string(got))
	//assert.Equal(t, expected, string(got))
}

type PointerComponentUser struct {
	Name string `json:"name"`
}

// AsComponent on a pointer type registers under the element type's name, and
// SchemaComponent registers a model after the fact without duplicating it.
func TestSchemaComponent_PointerAndExplicit(t *testing.T) {
	t.Parallel()

	doc, err := arrest.NewDocument("test")
	require.NoError(t, err)
	doc.PackageMap("test", "github.com/zostay/arrest-go_test")

	ptrModel := arrest.ModelFrom[*PointerComponentUser](doc, arrest.AsComponent())
	assert.Equal(t, "github.com/zostay/arrest-go_test.PointerComponentUser", ptrModel.Name)

	composed := arrest.OneOfTheseModels(doc, ptrModel, arrest.ModelFrom[PointerComponentUser](doc))
	doc.SchemaComponent("Either", composed)

	require.NoError(t, doc.Err())

	names := []string{}
	for _, sc := range doc.SchemaComponents(context.Background()) {
		names = append(names, sc.Name())
	}
	assert.ElementsMatch(t, []string{"test.PointerComponentUser", "Either"}, names)
}
