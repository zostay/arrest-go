package arrest_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zostay/arrest-go"
)

type ConnectionType int

type ListConnectionsResponse struct {
	Connections []*Connection `json:"connections" openapi:"connections,elemRefName=Connection"`
}

type CreateConnectionRequest struct {
	Connection Connection `json:"connection" openapi:"connection,refName=Connection"`
}

type CreateConnectionResponse struct {
	Connection Connection `json:"connection" openapi:",refName=Connection"`
}

type GetConnectionResponse struct {
	Connection Connection `json:"connection" openapi:",refName=Connection"`
}

type UpdateConnectionRequest struct {
	Connection Connection `json:"connection" openapi:",refName=Connection"`
}

type UpdateConnectionResponse struct {
	Connection Connection `json:"connection" openapi:",refName=Connection"`
}

type DeleteConnectionRequest struct {
	ID string `json:"id"`
}

type DeleteConnectionResponse struct{}

type Connection struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Type        ConnectionType      `json:"type" openapi:",type=string"`
	Properties  map[string][]string `json:"properties"`
	Secrets     map[string][]string `json:"secrets"`
}

type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func GetConnectionByID(context.Context, string) (*Connection, error) {
	return nil, nil
}

func DeleteConnection(context.Context, string) error {
	return nil
}

func OpenAPI(doc *arrest.Document) error {
	doc.Title("Connection Service").
		Description("The Connection Service provides CRUD operations for connection configurations.").
		Version("0.0.1")

	doc.PackageMap("zostay.arrest.test.v1", "github.com/zostay/arrest-go_test")
	doc.PackageMap("zostay.arrest.test.v1", "command-line-arguments_test")

	errModel := arrest.ModelFrom[ErrorPayload](doc, arrest.AsComponent()).
		Description("An error response.")
	errRef := arrest.SchemaRef(errModel.MappedName(doc.PkgMap))

	// List Connections
	{
		resModel := arrest.ModelFrom[ListConnectionsResponse](doc, arrest.AsComponent()).
			Description("The list of connection configurations.")
		resRef := arrest.SchemaRef(resModel.MappedName(doc.PkgMap))

		doc.Get("/connections").
			Description("List all connection configurations").
			Tags("Connections").
			Response("200", func(r *arrest.Response) {
				r.Content("application/json", resRef).
					Description("Returns the list of connection configurations.")
			}).
			Response("default", func(r *arrest.Response) {
				r.Content("application/json", errRef).
					Description("An error response if something went wrong.")
			})
	}

	// Create Connection
	{
		reqModel := arrest.ModelFrom[CreateConnectionRequest](doc, arrest.AsComponent()).
			Description("The request to create a new connection configuration.")
		reqRef := arrest.SchemaRef(reqModel.MappedName(doc.PkgMap))
		resModel := arrest.ModelFrom[CreateConnectionResponse](doc, arrest.AsComponent()).
			Description("The response to creating a new connection configuration.")
		resRef := arrest.SchemaRef(resModel.MappedName(doc.PkgMap))

		doc.Post("/connections").
			Description("Create a new connection configuration").
			Tags("Connections").
			RequestBody("application/json", reqRef).
			Response("200", func(r *arrest.Response) {
				r.Content("application/json", resRef).
					Description("Returns the connection configuration.")
			}).
			Response("default", func(r *arrest.Response) {
				r.Content("application/json", errRef).
					Description("An error response if something went wrong.")
			})
	}

	// Get Connection
	{
		resModel := arrest.ModelFrom[GetConnectionResponse](doc, arrest.AsComponent()).
			Description("The response to getting a connection configuration.")
		resRef := arrest.SchemaRef(resModel.MappedName(doc.PkgMap))

		getConnectionByID := arrest.ParametersFromReflect(reflect.TypeOf(GetConnectionByID)).
			P(0, func(p *arrest.Parameter) {
				p.Name("id").In("path").Required().
					Description("The ID of the connection configuration to retrieve")
			})

		doc.Get("/connections/{id}").
			Description("Get a connection configuration").
			Tags("Connections").
			Parameters(getConnectionByID).
			Response("200", func(r *arrest.Response) {
				r.Content("application/json", resRef).
					Description("Returns the connection configuration.")
			}).
			Response("default", func(r *arrest.Response) {
				r.Content("application/json", errRef).
					Description("An error response if something went wrong.")
			})
	}

	// Update Connection
	{
		reqModel := arrest.ModelFrom[UpdateConnectionRequest](doc, arrest.AsComponent()).
			Description("The request to update a connection configuration.")
		reqRef := arrest.SchemaRef(reqModel.MappedName(doc.PkgMap))

		resModel := arrest.ModelFrom[UpdateConnectionResponse](doc, arrest.AsComponent()).
			Description("The response to updating a connection configuration.")
		resRef := arrest.SchemaRef(resModel.MappedName(doc.PkgMap))

		updateConnectionWithId := arrest.NParameters(1).
			P(0, func(p *arrest.Parameter) {
				p.Name("id").In("path").Required().
					Model(arrest.ModelFrom[string](doc)).
					Description("The ID of the connection configuration to update")
			})

		doc.Put("/connections/{id}").
			Description("Update a connection configuration").
			Tags("Connections").
			Parameters(updateConnectionWithId).
			RequestBody("application/json", reqRef).
			Response("200", func(r *arrest.Response) {
				r.Content("application/json", resRef).
					Description("Returns the updated connection configuration.")
			}).
			Response("default", func(r *arrest.Response) {
				r.Content("application/json", errRef).
					Description("An error response if something went wrong.")
			})
	}

	// Delete Connection
	{
		deleteConnectionById := arrest.ParametersFromReflect(reflect.TypeOf(DeleteConnection)).
			P(0, func(p *arrest.Parameter) {
				p.Name("id").In("path").Required().
					Description("The ID of the connection configuration to delete")
			})

		doc.Delete("/connections/{id}").
			Description("Delete a connection configuration").
			Tags("Connections").
			Parameters(deleteConnectionById).
			Response("204", func(r *arrest.Response) {
				r.Description("The connection configuration was deleted.")
			}).
			Response("default", func(r *arrest.Response) {
				r.Content("application/json", errRef).
					Description("An error response if something went wrong.")
			})
	}

	return nil
}

const expect = `openapi: 3.1.0
info:
    title: Connection Service
    description: The Connection Service provides CRUD operations for connection configurations.
    version: 0.0.1
paths:
    /connections:
        get:
            tags:
                - Connections
            description: List all connection configurations
            responses:
                "200":
                    description: Returns the list of connection configurations.
                    content:
                        application/json:
                            schema:
                                $ref: '#/components/schemas/zostay.arrest.test.v1.ListConnectionsResponse'
                default:
                    description: An error response if something went wrong.
                    content:
                        application/json:
                            schema:
                                $ref: '#/components/schemas/zostay.arrest.test.v1.ErrorPayload'
        post:
            tags:
                - Connections
            description: Create a new connection configuration
            requestBody:
                content:
                    application/json:
                        schema:
                            $ref: '#/components/schemas/zostay.arrest.test.v1.CreateConnectionRequest'
            responses:
                "200":
                    description: Returns the connection configuration.
                    content:
                        application/json:
                            schema:
                                $ref: '#/components/schemas/zostay.arrest.test.v1.CreateConnectionResponse'
                default:
                    description: An error response if something went wrong.
                    content:
                        application/json:
                            schema:
                                $ref: '#/components/schemas/zostay.arrest.test.v1.ErrorPayload'
    /connections/{id}:
        get:
            tags:
                - Connections
            description: Get a connection configuration
            parameters:
                - name: id
                  in: path
                  description: The ID of the connection configuration to retrieve
                  required: true
                  schema:
                    type: string
            responses:
                "200":
                    description: Returns the connection configuration.
                    content:
                        application/json:
                            schema:
                                $ref: '#/components/schemas/zostay.arrest.test.v1.GetConnectionResponse'
                default:
                    description: An error response if something went wrong.
                    content:
                        application/json:
                            schema:
                                $ref: '#/components/schemas/zostay.arrest.test.v1.ErrorPayload'
        put:
            tags:
                - Connections
            description: Update a connection configuration
            parameters:
                - name: id
                  in: path
                  description: The ID of the connection configuration to update
                  required: true
                  schema:
                    type: string
            requestBody:
                content:
                    application/json:
                        schema:
                            $ref: '#/components/schemas/zostay.arrest.test.v1.UpdateConnectionRequest'
            responses:
                "200":
                    description: Returns the updated connection configuration.
                    content:
                        application/json:
                            schema:
                                $ref: '#/components/schemas/zostay.arrest.test.v1.UpdateConnectionResponse'
                default:
                    description: An error response if something went wrong.
                    content:
                        application/json:
                            schema:
                                $ref: '#/components/schemas/zostay.arrest.test.v1.ErrorPayload'
        delete:
            tags:
                - Connections
            description: Delete a connection configuration
            parameters:
                - name: id
                  in: path
                  description: The ID of the connection configuration to delete
                  required: true
                  schema:
                    type: string
            responses:
                "204":
                    description: The connection configuration was deleted.
                default:
                    description: An error response if something went wrong.
                    content:
                        application/json:
                            schema:
                                $ref: '#/components/schemas/zostay.arrest.test.v1.ErrorPayload'
components:
    schemas:
        zostay.arrest.test.v1.ErrorPayload:
            type: object
            properties:
                code:
                    type: string
                message:
                    type: string
            description: An error response.
        zostay.arrest.test.v1.ListConnectionsResponse:
            type: object
            properties:
                connections:
                    type: array
                    items:
                        $ref: '#/components/schemas/zostay.arrest.test.v1.Connection'
            description: The list of connection configurations.
        zostay.arrest.test.v1.Connection:
            type: object
            properties:
                id:
                    type: string
                name:
                    type: string
                description:
                    type: string
                type:
                    type: string
                properties:
                    type: object
                    additionalProperties:
                        type: array
                        items:
                            type: string
                secrets:
                    type: object
                    additionalProperties:
                        type: array
                        items:
                            type: string
        zostay.arrest.test.v1.CreateConnectionRequest:
            type: object
            properties:
                connection:
                    $ref: '#/components/schemas/zostay.arrest.test.v1.Connection'
            description: The request to create a new connection configuration.
        zostay.arrest.test.v1.CreateConnectionResponse:
            type: object
            properties:
                connection:
                    $ref: '#/components/schemas/zostay.arrest.test.v1.Connection'
            description: The response to creating a new connection configuration.
        zostay.arrest.test.v1.GetConnectionResponse:
            type: object
            properties:
                connection:
                    $ref: '#/components/schemas/zostay.arrest.test.v1.Connection'
            description: The response to getting a connection configuration.
        zostay.arrest.test.v1.UpdateConnectionRequest:
            type: object
            properties:
                connection:
                    $ref: '#/components/schemas/zostay.arrest.test.v1.Connection'
            description: The request to update a connection configuration.
        zostay.arrest.test.v1.UpdateConnectionResponse:
            type: object
            properties:
                connection:
                    $ref: '#/components/schemas/zostay.arrest.test.v1.Connection'
            description: The response to updating a connection configuration.
`

func TestDocument(t *testing.T) {
	t.Parallel()

	doc, err := arrest.NewDocument("")
	require.NotNil(t, doc)
	require.NoError(t, err)
	require.NoError(t, doc.Err())

	err = OpenAPI(doc)
	assert.NoError(t, err)

	rend, err := doc.OpenAPI.Render()
	assert.NoError(t, err)
	assert.NotEmpty(t, rend)
	assert.YAMLEq(t, expect, string(rend))
	//assert.Equal(t, expect, string(rend))
}

func TestDocumentSkipDocumentation(t *testing.T) {
	// global variables used, do not t.Parallel()
	arrest.SkipDocumentation = true
	defer func() { arrest.SkipDocumentation = false }()

	doc, err := arrest.NewDocument("")
	require.NotNil(t, doc)
	require.NoError(t, err)
	require.NoError(t, doc.Err())

	err = OpenAPI(doc)
	assert.NoError(t, err)

	rend, err := doc.OpenAPI.Render()
	assert.NoError(t, err)
	assert.NotEmpty(t, rend)
	assert.YAMLEq(t, expect, string(rend))
	//assert.Equal(t, expect, string(rend))
}

type ParameterTypeOverrideRequest struct {
	ID       string `json:"id" openapi:",in=path"`
	UserType int    `json:"userType" openapi:",in=query,type=string"`
	Count    uint64 `json:"count" openapi:",in=query,type=integer"`
}

func TestParametersFromStructWithTypeOverrides(t *testing.T) {
	t.Parallel()

	params := arrest.ParametersFrom[ParameterTypeOverrideRequest]()
	require.NoError(t, params.Err())
	require.Len(t, params.Parameters, 3)

	// Check ID parameter - should be path parameter with string type (inferred)
	idParam := params.Parameters[0]
	assert.Equal(t, "id", idParam.Parameter.Name)
	assert.Equal(t, "path", idParam.Parameter.In)
	assert.NotNil(t, idParam.Parameter.Required)
	assert.True(t, *idParam.Parameter.Required)
	assert.Equal(t, []string{"string"}, idParam.Parameter.Schema.Schema().Type)

	// Check UserType parameter - should be query parameter with string type (overridden from int)
	userTypeParam := params.Parameters[1]
	assert.Equal(t, "userType", userTypeParam.Parameter.Name)
	assert.Equal(t, "query", userTypeParam.Parameter.In)
	assert.Nil(t, userTypeParam.Parameter.Required) // Query params are not required by default
	assert.Equal(t, []string{"string"}, userTypeParam.Parameter.Schema.Schema().Type)

	// Check Count parameter - should be query parameter with integer type (overridden from uint64)
	countParam := params.Parameters[2]
	assert.Equal(t, "count", countParam.Parameter.Name)
	assert.Equal(t, "query", countParam.Parameter.In)
	assert.Nil(t, countParam.Parameter.Required) // Query params are not required by default
	assert.Equal(t, []string{"integer"}, countParam.Parameter.Schema.Schema().Type)
}
