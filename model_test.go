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

	doc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	myEnum := arrest.ModelFrom[string](doc).OneOf(
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
	doc.PackageMap(
		"zostay.test", "command-line-arguments_test",
	)

	model := arrest.ModelFrom[PersonWithRefName](doc, arrest.AsComponent())
	reqRef := arrest.SchemaRef(model.MappedName(doc.PkgMap))

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
	doc.PackageMap(
		"zostay.test", "command-line-arguments_test",
	)

	model := arrest.ModelFrom[PersonWithElemRefName](doc, arrest.AsComponent())
	reqRef := arrest.SchemaRef(model.MappedName(doc.PkgMap))

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
	doc.PackageMap(
		"zostay.test", "command-line-arguments_test",
	)

	// This test ensures that recursive structs don't cause infinite recursion
	model := arrest.ModelFrom[RecursiveStruct](doc, arrest.AsComponent())
	resRef := arrest.SchemaRef(model.MappedName(doc.PkgMap))

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
	doc.PackageMap(
		"zostay.test", "command-line-arguments_test",
	)

	// This test ensures that recursive structs don't cause infinite recursion
	model := arrest.ModelFrom[DeeperRecursiveStruct](doc, arrest.AsComponent())
	resRef := arrest.SchemaRef(model.MappedName(doc.PkgMap))

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

	doc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	// Test more complex recursive patterns
	model := arrest.ModelFrom[DeepRecursiveStruct](doc)
	assert.NoError(t, model.Err())

	require.NotNil(t, model.SchemaProxy)
	require.NotNil(t, model.SchemaProxy.Schema())

	// Should handle multiple levels of recursion
	refs := model.ExtractChildRefs()
	assert.NotEmpty(t, refs, "Should have child references for deeply recursive types")
}

type Account struct {
	Name     string              `json:"name"`
	Parent   *Account            `json:"parent,omitempty" openapi:",refName=Account"`
	Children map[string]*Account `json:"children,omitempty" openapi:",elemRefName=Account"`
}

type Commodity struct {
	Name string `json:"name"`
}

type Line struct {
	Description string     `json:"description"`
	Commodity   *Commodity `json:"commodity" openapi:",refName=Commodity"`
	Account     *Account   `json:"account,omitempty" openapi:",refName=Account"`
}

type Ledger struct {
	Description string     `json:"description"`
	Commodity   *Commodity `json:"commodity" openapi:",refName=Commodity"`
	Lines       []*Line    `json:"lines,omitempty" openapi:",elemRefName=Line"`
}

type LedgerRequest struct {
	Entry *Ledger `json:"entry,omitempty" openapi:",refName=Ledger"`
}

type LedgerResponse struct {
	Entry []*Ledger `json:"entry,omitempty" openapi:",elemRefName=Ledger"`
}

const expected_LedgerRequest = `openapi: 3.1.0
info:
  title: Ledger Request
paths:
  /ledger:
    get:
      parameters:
        - name: description
          in: query
          schema:
            type: string
          description: Filter by description (optional)
      responses:
        "200":
          description: Ledger Request Response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/zostay.test.LedgerResponse'
components:
  schemas:
    zostay.test.LedgerResponse:
      type: object
      properties:
        entry:
          type: array
          items:
            $ref: '#/components/schemas/zostay.test.Ledger'
    zostay.test.Ledger:
      type: object
      properties:
        description:
          type: string
        commodity:
          $ref: '#/components/schemas/zostay.test.Commodity'
        lines:
          type: array
          items:
            $ref: '#/components/schemas/zostay.test.Line'
    zostay.test.Line:
      type: object
      properties:
        description:
          type: string
        commodity:
          $ref: '#/components/schemas/zostay.test.Commodity'
        account:
          $ref: '#/components/schemas/zostay.test.Account'
    zostay.test.Commodity:
      type: object
      properties:
        name:
          type: string
    zostay.test.Account:
      type: object
      properties:
        name:
          type: string
        parent:
          $ref: '#/components/schemas/zostay.test.Account'
        children:
          type: object
          additionalProperties:
            $ref: '#/components/schemas/zostay.test.Account'
`

func TestModelFrom_Ledger(t *testing.T) {
	t.Parallel()

	doc, err := arrest.NewDocument("Ledger Request")
	require.NoError(t, err)

	doc.PackageMap(
		"zostay.test", "github.com/zostay/arrest-go_test",
	)
	doc.PackageMap(
		"zostay.test", "command-line-arguments_test",
	)

	listAccounts := arrest.NParameters(1).
		P(0, func(p *arrest.Parameter) {
			p.Name("description").In("query").
				Model(arrest.ModelFrom[string](doc)).
				Description("Filter by description (optional)")
		})

	ledgerModel := arrest.ModelFrom[LedgerResponse](doc, arrest.AsComponent())
	ledgerResRef := arrest.SchemaRef(ledgerModel.MappedName(doc.PkgMap))
	doc.Get("/ledger").
		Parameters(listAccounts).
		Response("200", func(r *arrest.Response) {
			r.Content("application/json", ledgerResRef).
				Description("Ledger Request Response")
		})
	assert.NoError(t, doc.Err())

	oas, err := doc.OpenAPI.Render()
	require.NoError(t, err)

	assert.YAMLEq(t, expected_LedgerRequest, string(oas))
	//assert.Equal(t, expected_LedgerRequest, string(oas))
}

// Test types for polymorphic tests
type Dog struct {
	PetType string `json:"petType"`
	Breed   string `json:"breed"`
	Name    string `json:"name"`
}

type Cat struct {
	PetType string `json:"petType"`
	Lives   int    `json:"lives"`
	Name    string `json:"name"`
}

type Bird struct {
	PetType string `json:"petType"`
	CanFly  bool   `json:"canFly"`
	Name    string `json:"name"`
}

const expected_OneOfTheseModels = `openapi: 3.1.0
info:
  title: test
paths:
  /pets:
    post:
      requestBody:
        content:
          application/json:
            schema:
              oneOf:
              - type: object
                properties:
                  petType:
                    type: string
                  breed:
                    type: string
                  name:
                    type: string
              - type: object
                properties:
                  petType:
                    type: string
                  lives:
                    type: integer
                    format: int32
                  name:
                    type: string
`

func TestOneOfTheseModels(t *testing.T) {
	t.Parallel()

	doc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	dogModel := arrest.ModelFrom[Dog](doc)
	catModel := arrest.ModelFrom[Cat](doc)

	petModel := arrest.OneOfTheseModels(doc, dogModel, catModel)
	assert.NoError(t, petModel.Err())

	doc.Post("/pets").
		RequestBody("application/json", petModel)

	assert.NoError(t, doc.Err())

	oas, err := doc.OpenAPI.Render()
	require.NoError(t, err)

	assert.YAMLEq(t, expected_OneOfTheseModels, string(oas))
}

const expected_AnyOfTheseModels = `openapi: 3.1.0
info:
  title: test
paths:
  /pets:
    post:
      requestBody:
        content:
          application/json:
            schema:
              anyOf:
              - type: object
                properties:
                  petType:
                    type: string
                  breed:
                    type: string
                  name:
                    type: string
              - type: object
                properties:
                  petType:
                    type: string
                  lives:
                    type: integer
                    format: int32
                  name:
                    type: string
`

func TestAnyOfTheseModels(t *testing.T) {
	t.Parallel()

	doc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	dogModel := arrest.ModelFrom[Dog](doc)
	catModel := arrest.ModelFrom[Cat](doc)

	petModel := arrest.AnyOfTheseModels(doc, dogModel, catModel)
	assert.NoError(t, petModel.Err())

	doc.Post("/pets").
		RequestBody("application/json", petModel)

	assert.NoError(t, doc.Err())

	oas, err := doc.OpenAPI.Render()
	require.NoError(t, err)

	assert.YAMLEq(t, expected_AnyOfTheseModels, string(oas))
}

const expected_AllOfTheseModels = `openapi: 3.1.0
info:
  title: test
paths:
  /pets:
    post:
      requestBody:
        content:
          application/json:
            schema:
              allOf:
              - type: object
                properties:
                  petType:
                    type: string
                  breed:
                    type: string
                  name:
                    type: string
              - type: object
                properties:
                  petType:
                    type: string
                  lives:
                    type: integer
                    format: int32
                  name:
                    type: string
`

func TestAllOfTheseModels(t *testing.T) {
	t.Parallel()

	doc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	dogModel := arrest.ModelFrom[Dog](doc)
	catModel := arrest.ModelFrom[Cat](doc)

	hybridModel := arrest.AllOfTheseModels(doc, dogModel, catModel)
	assert.NoError(t, hybridModel.Err())

	doc.Post("/pets").
		RequestBody("application/json", hybridModel)

	assert.NoError(t, doc.Err())

	oas, err := doc.OpenAPI.Render()
	require.NoError(t, err)

	assert.YAMLEq(t, expected_AllOfTheseModels, string(oas))
}

const expected_DiscriminatorOneOf = `openapi: 3.1.0
info:
  title: test
paths:
  /pets:
    post:
      requestBody:
        content:
          application/json:
            schema:
              oneOf:
              - type: object
                properties:
                  petType:
                    type: string
                  breed:
                    type: string
                  name:
                    type: string
              - type: object
                properties:
                  petType:
                    type: string
                  lives:
                    type: integer
                    format: int32
                  name:
                    type: string
              - type: object
                properties:
                  petType:
                    type: string
                  canFly:
                    type: boolean
                  name:
                    type: string
              discriminator:
                propertyName: petType
                defaultMapping: dog
                mapping:
                  dog: '#/components/schemas/github.com.zostay.arrest-go_test.Dog'
                  cat: '#/components/schemas/github.com.zostay.arrest-go_test.Cat'
                  bird: '#/components/schemas/github.com.zostay.arrest-go_test.Bird'
components:
  schemas:
    github.com.zostay.arrest-go_test.Dog:
      type: object
      properties:
        petType:
          type: string
        breed:
          type: string
        name:
          type: string
    github.com.zostay.arrest-go_test.Cat:
      type: object
      properties:
        petType:
          type: string
        lives:
          type: integer
          format: int32
        name:
          type: string
    github.com.zostay.arrest-go_test.Bird:
      type: object
      properties:
        petType:
          type: string
        canFly:
          type: boolean
        name:
          type: string
`

func TestDiscriminatorWithOneOf(t *testing.T) {
	t.Parallel()

	doc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	dogModel := arrest.ModelFrom[Dog](doc, arrest.AsComponent())
	catModel := arrest.ModelFrom[Cat](doc, arrest.AsComponent())
	birdModel := arrest.ModelFrom[Bird](doc, arrest.AsComponent())

	petModel := arrest.OneOfTheseModels(doc, dogModel, catModel, birdModel).
		Discriminator("petType", "dog",
			"dog", "#/components/schemas/"+dogModel.MappedName(doc.PkgMap),
			"cat", "#/components/schemas/"+catModel.MappedName(doc.PkgMap),
			"bird", "#/components/schemas/"+birdModel.MappedName(doc.PkgMap))

	assert.NoError(t, petModel.Err())

	doc.Post("/pets").
		RequestBody("application/json", petModel)

	assert.NoError(t, doc.Err())

	oas, err := doc.OpenAPI.Render()
	require.NoError(t, err)

	assert.YAMLEq(t, expected_DiscriminatorOneOf, string(oas))
}

func TestDiscriminatorInvalidMappings(t *testing.T) {
	t.Parallel()

	doc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	dogModel := arrest.ModelFrom[Dog](doc)
	catModel := arrest.ModelFrom[Cat](doc)

	// Test odd number of mapping arguments (should cause error)
	petModel := arrest.OneOfTheseModels(doc, dogModel, catModel).
		Discriminator("petType", "dog", "dog") // Missing the value for "dog"

	assert.Error(t, petModel.Err())
	assert.Contains(t, petModel.Err().Error(), "discriminator mappings must be provided in pairs")
}

func TestEmptyModelsError(t *testing.T) {
	t.Parallel()

	doc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	// Test OneOfTheseModels with no models
	oneOfModel := arrest.OneOfTheseModels(doc)
	assert.Error(t, oneOfModel.Err())
	assert.ErrorIs(t, oneOfModel.Err(), arrest.ErrUnsupportedModelType)

	// Test AnyOfTheseModels with no models
	anyOfModel := arrest.AnyOfTheseModels(doc)
	assert.Error(t, anyOfModel.Err())
	assert.ErrorIs(t, anyOfModel.Err(), arrest.ErrUnsupportedModelType)

	// Test AllOfTheseModels with no models
	allOfModel := arrest.AllOfTheseModels(doc)
	assert.Error(t, allOfModel.Err())
	assert.ErrorIs(t, allOfModel.Err(), arrest.ErrUnsupportedModelType)
}
