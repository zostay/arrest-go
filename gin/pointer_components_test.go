package gin

import (
	"context"
	"testing"

	ginHTTP "github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zostay/arrest-go"
)

// Test types for pointer components (using different names to avoid conflicts)
type CreateDogRequest struct {
	Name  string `json:"name"`
	Breed string `json:"breed"`
}

type Dog struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Breed string `json:"breed"`
}

// Controller function with pointer request and response types
func CreateDog(ctx context.Context, req *CreateDogRequest) (*Dog, error) {
	return &Dog{
		ID:    123,
		Name:  req.Name,
		Breed: req.Breed,
	}, nil
}

func TestCallWithPointerTypesAndComponents(t *testing.T) {
	t.Parallel()

	arrestDoc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	router := ginHTTP.New()
	doc := NewDocument(arrestDoc, router)

	// Test with pointer request and response types using WithComponents
	doc.Post("/dogs").
		OperationID("createDog").
		Call(CreateDog, WithComponents())

	require.NoError(t, arrestDoc.Err())

	// Render the OpenAPI spec
	oas, err := arrestDoc.OpenAPI.Render()
	require.NoError(t, err)
	spec := string(oas)

	// Check that request body uses component reference
	assert.Contains(t, spec, "$ref: '#/components/schemas/", "Request body should use component reference")
	assert.Contains(t, spec, "CreateDogRequest", "Should contain CreateDogRequest component")

	// Check that response uses component reference
	assert.Contains(t, spec, "Dog", "Should contain Dog component")

	// Verify components section exists and contains our types
	assert.Contains(t, spec, "components:", "Should have components section")
	assert.Contains(t, spec, "schemas:", "Should have schemas in components")

	// Check for specific component definitions
	assert.Contains(t, spec, "CreateDogRequest:", "Should define CreateDogRequest component")
	assert.Contains(t, spec, "Dog:", "Should define Dog component")

	// Verify the request body references the component
	assert.Contains(t, spec, "requestBody:", "Should have request body")

	// Verify the response references the component
	assert.Contains(t, spec, "responses:", "Should have responses")
	assert.Contains(t, spec, "\"200\":", "Should have 200 response")
}

func TestCallWithPointerTypesIndividualComponentOptions(t *testing.T) {
	t.Parallel()

	arrestDoc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	router := ginHTTP.New()
	doc := NewDocument(arrestDoc, router)

	// Test with individual component options
	doc.Post("/dogs-individual").
		OperationID("createDogWithIndividualOptions").
		Call(CreateDog,
			WithRequestComponent(),
			WithResponseComponent())

	require.NoError(t, arrestDoc.Err())

	// Render the OpenAPI spec
	oas, err := arrestDoc.OpenAPI.Render()
	require.NoError(t, err)
	spec := string(oas)

	// Same checks as above
	assert.Contains(t, spec, "$ref: '#/components/schemas/", "Request body should use component reference")
	assert.Contains(t, spec, "CreateDogRequest", "Should contain CreateDogRequest component")
	assert.Contains(t, spec, "Dog", "Should contain Dog component")
	assert.Contains(t, spec, "components:", "Should have components section")
	assert.Contains(t, spec, "CreateDogRequest:", "Should define CreateDogRequest component")
	assert.Contains(t, spec, "Dog:", "Should define Dog component")
}

func TestCallWithNonPointerTypesAndComponents(t *testing.T) {
	t.Parallel()

	// Controller function with non-pointer types for comparison
	controllerNonPointer := func(ctx context.Context, req CreateDogRequest) (Dog, error) {
		return Dog{
			ID:    123,
			Name:  req.Name,
			Breed: req.Breed,
		}, nil
	}

	arrestDoc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	router := ginHTTP.New()
	doc := NewDocument(arrestDoc, router)

	// Test with non-pointer request and response types using WithComponents
	doc.Post("/dogs-nonpointer").
		OperationID("createDogNonPointer").
		Call(controllerNonPointer, WithComponents())

	require.NoError(t, arrestDoc.Err())

	// Render the OpenAPI spec
	oas, err := arrestDoc.OpenAPI.Render()
	require.NoError(t, err)
	spec := string(oas)

	// Check that components are still created correctly for non-pointer types
	assert.Contains(t, spec, "CreateDogRequest", "Should contain CreateDogRequest component")
	assert.Contains(t, spec, "Dog", "Should contain Dog component")
}
