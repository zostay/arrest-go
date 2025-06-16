package arrest_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zostay/arrest-go"
)

const expected_WithOneOf = `openapi: 3.1.0
info:
  title: test
paths:
  /simple:
    put:
      requestBody:
        content:
          text/plain:
            schema:
              type: string
              oneOf:
              - title: Foo
                const: foo
                description: Foo is a test value
              - const: bar
                title: Bar
                description: Bar is a test value
`

func TestModelFrom_WithOneOf(t *testing.T) {
	t.Parallel()

	myEnum := arrest.ModelFrom[string]().OneOf(
		arrest.Enumeration{
			Const:       "foo",
			Title:       "Foo",
			Description: "Foo is a test value",
		},
		arrest.Enumeration{
			Const:       "bar",
			Title:       "Bar",
			Description: "Bar is a test value",
		},
	)

	assert.NoError(t, myEnum.Err())

	doc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	doc.Put("/simple").
		RequestBody("text/plain", myEnum)

	assert.NoError(t, doc.Err())

	oas, err := doc.OpenAPI.Render()
	require.NoError(t, err)

	assert.YAMLEq(t, expected_WithOneOf, string(oas))
	//assert.Equal(t, expected_WithOneOf, string(oas))
}

const expected_RefName = `openapi: 3.1.0
info:
  title: test
paths:
  /simple:
    put:
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/zostay.test.PersonWithRefName'
components:
  schemas:
    zostay.test.PersonWithRefName:
      type: object
      properties:
        name:
          type: string
        address:
          $ref: '#/components/schemas/zostay.test.Address'
    zostay.test.Address:
      type: object
      properties:
        street:
          type: string
        city:
          type: string
`

type Address struct {
	Street string `json:"street"`
	City   string `json:"city"`
}

type PersonWithRefName struct {
	Name    string  `json:"name"`
	Address Address `json:"address" openapi:"address,refName=Address"`
}

func TestModelFrom_RefName(t *testing.T) {
	t.Parallel()

	doc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	doc.PackageMap(
		"zostay.test", "github.com/zostay/arrest-go_test",
	)

	reqRef := doc.SchemaComponentRef(arrest.ModelFrom[PersonWithRefName]()).Ref()

	doc.Put("/simple").
		RequestBody("application/json", reqRef)

	assert.NoError(t, doc.Err())

	oas, err := doc.OpenAPI.Render()
	require.NoError(t, err)

	assert.YAMLEq(t, expected_RefName, string(oas))
	//assert.Equal(t, expected_RefName, string(oas))
}

const expected_ElemRefName = `openapi: 3.1.0
info:
  title: test
paths:
  /simple:
    put:
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/zostay.test.PersonWithElemRefName'
components:
  schemas:
    zostay.test.PersonWithElemRefName:
      type: object
      properties:
        name:
          type: string
        addresses:
          type: array
          items:
              $ref: '#/components/schemas/zostay.test.Address'
    zostay.test.Address:
      type: object
      properties:
        street:
          type: string
        city:
          type: string
`

type PersonWithElemRefName struct {
	Name      string    `json:"name"`
	Addresses []Address `json:"addresses" openapi:"addresses,elemRefName=Address"`
}

func TestModelFrom_ElemRefName(t *testing.T) {
	t.Parallel()

	doc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	doc.PackageMap(
		"zostay.test", "github.com/zostay/arrest-go_test",
	)

	reqRef := doc.SchemaComponentRef(arrest.ModelFrom[PersonWithElemRefName]()).Ref()

	doc.Put("/simple").
		RequestBody("application/json", reqRef)

	assert.NoError(t, doc.Err())

	oas, err := doc.OpenAPI.Render()
	require.NoError(t, err)

	assert.YAMLEq(t, expected_ElemRefName, string(oas))
	//assert.Equal(t, expected_ElemRefName, string(oas))
}

const expected_RecursiveStruct = `openapi: 3.1.0
info:
  title: test-recursive
paths:
  /recursive:
    get:
      responses:
        '200':
          description: the recursive struct is a recursive struct
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/zostay.test.RecursiveStruct'
components:
  schemas:
    zostay.test.RecursiveStruct:
      type: object
      properties:
        name:
          type: string
        children:
          type: array
          items:
            $ref: '#/components/schemas/zostay.test.RecursiveStruct'
        parent:
          $ref: '#/components/schemas/zostay.test.RecursiveStruct'
`

type RecursiveStruct struct {
	Name     string            `json:"name"`
	Children []RecursiveStruct `json:"children,omitempty"`
	Parent   *RecursiveStruct  `json:"parent,omitempty"`
}

func TestModelFrom_RecursiveStruct(t *testing.T) {
	t.Parallel()

	// Verify we can create a document with this recursive model
	doc, err := arrest.NewDocument("test-recursive")
	require.NoError(t, err)

	doc.PackageMap(
		"zostay.test", "github.com/zostay/arrest-go_test",
	)

	// This test ensures that recursive structs don't cause infinite recursion
	resRef := doc.SchemaComponentRef(arrest.ModelFrom[RecursiveStruct]()).Ref()

	doc.Get("/recursive").
		Response("200", func(r *arrest.Response) {
			r.Content("application/json", resRef).
				Description("the recursive struct is a recursive struct")
		})

	assert.NoError(t, doc.Err())

	// Should be able to render without hanging
	oas, err := doc.OpenAPI.Render()
	require.NoError(t, err)
	assert.NotEmpty(t, oas)

	assert.YAMLEq(t, expected_RecursiveStruct, string(oas))
	//assert.Equal(t, expected_RecursiveStruct, string(oas))
}

const expected_DeeperRecursiveStruct = `openapi: 3.1.0
info:
  title: test-deeper-recursive
paths:
  /recursive:
    get:
      responses:
        '200':
          description: the recursive struct is a recursive struct
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/zostay.test.DeeperRecursiveStruct'
components:
  schemas:
    zostay.test.DeeperRecursiveStruct:
      type: object
      properties:
        recursive:
          $ref: '#/components/schemas/zostay.test.RecursiveStruct'
    zostay.test.RecursiveStruct:
      type: object
      properties:
        name:
          type: string
        children:
          type: array
          items:
            $ref: '#/components/schemas/zostay.test.RecursiveStruct'
        parent:
          $ref: '#/components/schemas/zostay.test.RecursiveStruct'
`

type DeeperRecursiveStruct struct {
	Recursive RecursiveStruct `json:"recursive" openapi:"recursive,refName=RecursiveStruct"`
}

func TestModelFrom_DeeperRecursiveStruct(t *testing.T) {
	t.Parallel()

	// Verify we can create a document with this recursive model
	doc, err := arrest.NewDocument("test-deeper-recursive")
	require.NoError(t, err)

	doc.PackageMap(
		"zostay.test", "github.com/zostay/arrest-go_test",
	)

	// This test ensures that recursive structs don't cause infinite recursion
	resRef := doc.SchemaComponentRef(arrest.ModelFrom[DeeperRecursiveStruct]()).Ref()

	doc.Get("/recursive").
		Response("200", func(r *arrest.Response) {
			r.Content("application/json", resRef).
				Description("the recursive struct is a recursive struct")
		})

	assert.NoError(t, doc.Err())

	// Should be able to render without hanging
	oas, err := doc.OpenAPI.Render()
	require.NoError(t, err)
	assert.NotEmpty(t, oas)

	assert.YAMLEq(t, expected_DeeperRecursiveStruct, string(oas))
	//assert.Equal(t, expected_DeeperRecursiveStruct, string(oas))
}

type DeepRecursiveStruct struct {
	ID       int                             `json:"id"`
	Name     string                          `json:"name"`
	Children []*DeepRecursiveStruct          `json:"children,omitempty"`
	Sibling  *DeepRecursiveStruct            `json:"sibling,omitempty"`
	Meta     map[string]*DeepRecursiveStruct `json:"meta,omitempty"`
}

func TestModelFrom_DeepRecursiveStruct(t *testing.T) {
	t.Parallel()

	// Test more complex recursive patterns
	model := arrest.ModelFrom[DeepRecursiveStruct]()
	assert.NoError(t, model.Err())

	require.NotNil(t, model.SchemaProxy)
	require.NotNil(t, model.SchemaProxy.Schema())

	// Should handle multiple levels of recursion
	refs := model.ExtractChildRefs()
	assert.NotEmpty(t, refs, "Should have child references for deeply recursive types")
}
