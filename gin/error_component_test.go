package gin

import (
	"context"
	"testing"

	ginHTTP "github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zostay/arrest-go"
)

// ComponentAPIError is a custom error body used to test error component registration.
type ComponentAPIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ComponentJot struct {
	ID string `json:"id"`
}

func listComponentJots(ctx context.Context, input struct{}) ([]ComponentJot, error) {
	return nil, nil
}

func getComponentJot(ctx context.Context, input struct {
	ID string `json:"id" openapi:",in=path"`
}) (*ComponentJot, error) {
	return nil, nil
}

const expected_ErrorComponents = `openapi: 3.1.0
info:
  title: test
paths:
  /jots:
    get:
      operationId: listJots
      responses:
        "200":
          description: Success
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/test.ComponentJot'
        default:
          description: unexpected error
          content:
            application/json:
              schema:
                oneOf:
                  - $ref: '#/components/schemas/test.ErrorResponse'
                  - $ref: '#/components/schemas/test.ComponentAPIError'
  /jots/{id}:
    get:
      operationId: getJot
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        "200":
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/test.ComponentJot'
        default:
          description: unexpected error
          content:
            application/json:
              schema:
                oneOf:
                  - $ref: '#/components/schemas/test.ErrorResponse'
                  - $ref: '#/components/schemas/test.ComponentAPIError'
components:
  schemas:
    test.ComponentJot:
      type: object
      properties:
        id:
          type: string
      required:
        - id
    test.ErrorResponse:
      type: object
      description: ErrorResponse represents the standard error response format.
      properties:
        status:
          type: string
        type:
          type: string
        message:
          type: string
        fields:
          type: object
          additionalProperties:
            type: string
      required:
        - status
        - type
        - message
    test.ComponentAPIError:
      type: object
      properties:
        code:
          type: string
        message:
          type: string
      required:
        - code
        - message
`

// WithComponents registers the default ErrorResponse and custom error models
// as components and references them from every operation's default response.
func TestCallMethod_ErrorComponents(t *testing.T) {
	t.Parallel()

	arrestDoc, err := arrest.NewDocument("test")
	require.NoError(t, err)
	arrestDoc.PackageMap("test", "github.com/zostay/arrest-go/gin")

	router := ginHTTP.New()
	doc := NewDocument(arrestDoc, router)

	errModel := arrest.ModelFrom[ComponentAPIError](arrestDoc)

	doc.Get("/jots").OperationID("listJots").
		Call(listComponentJots, WithCallErrorModel(errModel), WithComponents())
	doc.Get("/jots/{id}").OperationID("getJot").
		Call(getComponentJot, WithCallErrorModel(errModel), WithComponents())

	require.NoError(t, arrestDoc.Err())
	oas, err := arrestDoc.OpenAPI.Render()
	require.NoError(t, err)
	assert.YAMLEq(t, expected_ErrorComponents, string(oas))
}

// WithErrorComponent on its own registers only the error models, leaving the
// request and response schemas inline.
func TestCallMethod_WithErrorComponentOnly(t *testing.T) {
	t.Parallel()

	arrestDoc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	router := ginHTTP.New()
	doc := NewDocument(arrestDoc, router)

	doc.Get("/jots").OperationID("listJots").
		Call(listComponentJots, WithErrorComponent())

	require.NoError(t, arrestDoc.Err())
	oas, err := arrestDoc.OpenAPI.Render()
	require.NoError(t, err)
	spec := string(oas)

	assert.Contains(t, spec, "$ref: '#/components/schemas/ErrorResponse'")
	assert.Contains(t, spec, "\n    ErrorResponse:\n")
	assert.NotContains(t, spec, "ComponentJot:")
	assert.NotContains(t, spec, "oneOf:")
}

// ReplaceCallErrorModel models are registered too, and a SchemaRef passed as
// an error model is used as-is.
func TestCallMethod_ErrorComponentReplaceAndRef(t *testing.T) {
	t.Parallel()

	arrestDoc, err := arrest.NewDocument("test")
	require.NoError(t, err)
	arrestDoc.PackageMap("test", "github.com/zostay/arrest-go/gin")

	router := ginHTTP.New()
	doc := NewDocument(arrestDoc, router)

	composed := arrest.OneOfTheseModels(arrestDoc,
		arrest.ModelFrom[ComponentAPIError](arrestDoc),
		arrest.ModelFrom[ErrorResponse](arrestDoc))
	arrestDoc.SchemaComponent("AnyError", composed)

	handler := func(ctx *ginHTTP.Context, err error) interface{} { return err }
	doc.Get("/jots").OperationID("listJots").
		Call(listComponentJots,
			ReplaceCallErrorModel(arrest.ModelFrom[ComponentAPIError](arrestDoc)),
			ReplaceCallErrorModel(arrest.SchemaRef("AnyError")),
			WithErrorHandler(handler),
			WithErrorComponent())

	require.NoError(t, arrestDoc.Err())
	oas, err := arrestDoc.OpenAPI.Render()
	require.NoError(t, err)
	spec := string(oas)

	assert.Contains(t, spec, "$ref: '#/components/schemas/test.ComponentAPIError'")
	assert.Contains(t, spec, "$ref: '#/components/schemas/AnyError'")
	assert.NotContains(t, spec, "ErrorResponse'")
}

// A composed model cannot be named, so asking to register it is an error.
func TestCallMethod_ErrorComponentUnnamedModel(t *testing.T) {
	t.Parallel()

	arrestDoc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	router := ginHTTP.New()
	doc := NewDocument(arrestDoc, router)

	composed := arrest.OneOfTheseModels(arrestDoc,
		arrest.ModelFrom[ComponentAPIError](arrestDoc),
		arrest.ModelFrom[ErrorResponse](arrestDoc))

	op := doc.Get("/jots").OperationID("listJots").
		Call(listComponentJots, WithCallErrorModel(composed), WithErrorComponent())

	require.ErrorContains(t, op.Err(), "not a named type")
}
