package gin

import (
	"context"
	"testing"

	ginHTTP "github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zostay/arrest-go"
)

// CustomErrorResponse for testing error model combination
type CustomAPIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func testControllerForErrorCombination(ctx context.Context, input struct{}) (*struct{}, error) {
	return &struct{}{}, nil
}

func TestWithCallErrorModel_CombinesWithDefaultErrorResponse(t *testing.T) {
	t.Parallel()

	arrestDoc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	router := ginHTTP.New()
	doc := NewDocument(arrestDoc, router)

	// Create a custom error model
	customErrorModel := arrest.ModelFrom[CustomAPIError](arrestDoc)

	// Configure operation with custom error model
	doc.Post("/test").
		OperationID("testOp").
		Call(testControllerForErrorCombination, WithCallErrorModel(customErrorModel))

	require.NoError(t, arrestDoc.Err())

	// Render the OpenAPI spec
	oas, err := arrestDoc.OpenAPI.Render()
	require.NoError(t, err)
	spec := string(oas)

	// Verify that both ErrorResponse and CustomAPIError are in the spec
	// The spec should contain a oneOf with both error types
	assert.Contains(t, spec, "oneOf:", "Should have oneOf for multiple error models")

	// Check for ErrorResponse properties (default error response)
	assert.Contains(t, spec, "status:", "Should contain ErrorResponse status field")
	assert.Contains(t, spec, "type:", "Should contain ErrorResponse type field")
	assert.Contains(t, spec, "message:", "Should contain ErrorResponse message field")

	// Check for CustomAPIError properties (custom error response)
	assert.Contains(t, spec, "code:", "Should contain CustomAPIError code field")
	assert.Contains(t, spec, "details:", "Should contain CustomAPIError details field")

	// Verify the default response references both error types
	assert.Contains(t, spec, "default:", "Should have default error response")
}

func TestWithCallErrorModel_MultipleCustomErrors(t *testing.T) {
	t.Parallel()

	arrestDoc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	router := ginHTTP.New()
	doc := NewDocument(arrestDoc, router)

	// Create multiple custom error models
	type ValidationError struct {
		Field   string `json:"field"`
		Message string `json:"message"`
	}

	type AuthError struct {
		Reason string `json:"reason"`
		Code   int    `json:"code"`
	}

	validationErrorModel := arrest.ModelFrom[ValidationError](arrestDoc)
	authErrorModel := arrest.ModelFrom[AuthError](arrestDoc)

	// Configure operation with multiple custom error models
	doc.Post("/test").
		OperationID("testMultipleErrors").
		Call(testControllerForErrorCombination,
			WithCallErrorModel(validationErrorModel),
			WithCallErrorModel(authErrorModel))

	require.NoError(t, arrestDoc.Err())

	// Render the OpenAPI spec
	oas, err := arrestDoc.OpenAPI.Render()
	require.NoError(t, err)
	spec := string(oas)

	// Should have oneOf with all three error types: ErrorResponse, ValidationError, AuthError
	assert.Contains(t, spec, "oneOf:", "Should have oneOf for multiple error models")

	// Check for default ErrorResponse
	assert.Contains(t, spec, "status:", "Should contain default ErrorResponse")

	// Check for ValidationError
	assert.Contains(t, spec, "field:", "Should contain ValidationError field")

	// Check for AuthError
	assert.Contains(t, spec, "reason:", "Should contain AuthError reason field")
}

func TestNoCustomErrorModel_UsesDefaultOnly(t *testing.T) {
	t.Parallel()

	arrestDoc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	router := ginHTTP.New()
	doc := NewDocument(arrestDoc, router)

	// Configure operation without custom error models
	doc.Post("/test").
		OperationID("testDefaultOnly").
		Call(testControllerForErrorCombination)

	require.NoError(t, arrestDoc.Err())

	// Render the OpenAPI spec
	oas, err := arrestDoc.OpenAPI.Render()
	require.NoError(t, err)
	spec := string(oas)

	// Should NOT have oneOf since there's only one error type
	assert.NotContains(t, spec, "oneOf:", "Should not have oneOf for single error model")

	// Should contain default ErrorResponse properties
	assert.Contains(t, spec, "status:", "Should contain ErrorResponse status field")
	assert.Contains(t, spec, "type:", "Should contain ErrorResponse type field")
	assert.Contains(t, spec, "message:", "Should contain ErrorResponse message field")
}
