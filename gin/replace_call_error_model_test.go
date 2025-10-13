package gin

import (
	"context"
	"errors"
	"net/http"
	"testing"

	ginHTTP "github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zostay/arrest-go"
)

// Custom error types for testing ReplaceCallErrorModel
type APIError struct {
	ErrorCode string `json:"error_code"`
	Details   string `json:"details"`
}

func (e *APIError) Error() string {
	return e.Details
}

func (e *APIError) HTTPStatusCode() int {
	return http.StatusBadRequest
}

type SystemError struct {
	Component string `json:"component"`
	Message   string `json:"message"`
}

func (e *SystemError) Error() string {
	return e.Message
}

func (e *SystemError) HTTPStatusCode() int {
	return http.StatusServiceUnavailable
}

// Test controller functions
func controllerForReplaceErrorModel(ctx context.Context, input struct{}) (*struct{}, error) {
	return &struct{}{}, nil
}

func controllerThatReturnsError(ctx context.Context, input struct{}) (*struct{}, error) {
	return nil, errors.New("controller error")
}

func controllerThatPanics(ctx context.Context, input struct{}) (*struct{}, error) {
	panic("something went wrong")
}

func TestReplaceCallErrorModel_MutualExclusivity(t *testing.T) {
	t.Parallel()

	arrestDoc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	router := ginHTTP.New()
	doc := NewDocument(arrestDoc, router)

	customErrorModel := arrest.ModelFrom[APIError](arrestDoc)

	// Test that ReplaceCallErrorModel and WithCallErrorModel are mutually exclusive
	operation := doc.Post("/test").Call(controllerForReplaceErrorModel,
		WithCallErrorModel(customErrorModel),
		ReplaceCallErrorModel(customErrorModel))

	// Should have error due to mutual exclusivity
	assert.Error(t, operation.Err())
	assert.Contains(t, operation.Err().Error(), "mutually exclusive")
}

func TestReplaceCallErrorModel_RequiresErrorHandler(t *testing.T) {
	t.Parallel()

	arrestDoc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	router := ginHTTP.New()
	doc := NewDocument(arrestDoc, router)

	customErrorModel := arrest.ModelFrom[APIError](arrestDoc)

	// Test that ReplaceCallErrorModel requires WithErrorHandler
	operation := doc.Post("/test").Call(controllerForReplaceErrorModel,
		ReplaceCallErrorModel(customErrorModel))

	// Should have error due to missing error handler
	assert.Error(t, operation.Err())
	assert.Contains(t, operation.Err().Error(), "requires WithErrorHandler")
}

func TestReplaceCallErrorModel_SingleErrorModel(t *testing.T) {
	t.Parallel()

	arrestDoc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	router := ginHTTP.New()
	doc := NewDocument(arrestDoc, router)

	customErrorModel := arrest.ModelFrom[APIError](arrestDoc)

	customErrorHandler := func(ctx *ginHTTP.Context, err error) interface{} {
		return &APIError{
			ErrorCode: "CUSTOM_ERROR",
			Details:   err.Error(),
		}
	}

	// Configure operation with ReplaceCallErrorModel
	doc.Post("/test").
		OperationID("testReplaceErrorModel").
		Call(controllerForReplaceErrorModel,
			ReplaceCallErrorModel(customErrorModel),
			WithErrorHandler(customErrorHandler))

	require.NoError(t, arrestDoc.Err())

	// Render the OpenAPI spec
	oas, err := arrestDoc.OpenAPI.Render()
	require.NoError(t, err)
	spec := string(oas)

	// Should NOT contain default ErrorResponse properties
	assert.NotContains(t, spec, "status:", "Should not contain default ErrorResponse status field")

	// Should contain custom error properties
	assert.Contains(t, spec, "error_code:", "Should contain APIError error_code field")
	assert.Contains(t, spec, "details:", "Should contain APIError details field")

	// Should NOT have oneOf since there's only one error model
	assert.NotContains(t, spec, "oneOf:", "Should not have oneOf for single error model")
}

func TestReplaceCallErrorModel_MultipleErrorModels(t *testing.T) {
	t.Parallel()

	arrestDoc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	router := ginHTTP.New()
	doc := NewDocument(arrestDoc, router)

	apiErrorModel := arrest.ModelFrom[APIError](arrestDoc)
	systemErrorModel := arrest.ModelFrom[SystemError](arrestDoc)

	customErrorHandler := func(ctx *ginHTTP.Context, err error) interface{} {
		return &APIError{
			ErrorCode: "CUSTOM_ERROR",
			Details:   err.Error(),
		}
	}

	// Configure operation with multiple ReplaceCallErrorModel
	doc.Post("/test").
		OperationID("testMultipleReplaceErrorModels").
		Call(controllerForReplaceErrorModel,
			ReplaceCallErrorModel(apiErrorModel),
			ReplaceCallErrorModel(systemErrorModel),
			WithErrorHandler(customErrorHandler))

	require.NoError(t, arrestDoc.Err())

	// Render the OpenAPI spec
	oas, err := arrestDoc.OpenAPI.Render()
	require.NoError(t, err)
	spec := string(oas)

	// Should have oneOf with multiple custom error types
	assert.Contains(t, spec, "oneOf:", "Should have oneOf for multiple error models")

	// Should NOT contain default ErrorResponse properties
	assert.NotContains(t, spec, "status:", "Should not contain default ErrorResponse status field")

	// Should contain custom error properties
	assert.Contains(t, spec, "error_code:", "Should contain APIError error_code field")
	assert.Contains(t, spec, "component:", "Should contain SystemError component field")
}

func TestReplaceCallErrorModel_CustomErrorHandlerInAllScenarios(t *testing.T) {
	tests := []struct {
		name       string
		controller interface{}
		setupFunc  func(*ginHTTP.Context)
		options    []CallOption
	}{
		{
			name:       "Controller Error",
			controller: controllerThatReturnsError,
			setupFunc:  nil,
		},
		{
			name:       "Panic Protection",
			controller: controllerThatPanics,
			setupFunc:  nil,
			options:    []CallOption{WithPanicProtection()},
		},
		{
			name:       "Validation Error",
			controller: controllerForReplaceErrorModel,
			setupFunc: func(c *ginHTTP.Context) {
				// This will cause extractInput to fail by providing invalid JSON
				c.Request.Header.Set("Content-Type", "application/json")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			arrestDoc, err := arrest.NewDocument("test")
			require.NoError(t, err)

			router := ginHTTP.New()
			doc := NewDocument(arrestDoc, router)

			customErrorModel := arrest.ModelFrom[APIError](arrestDoc)

			customErrorHandler := func(ctx *ginHTTP.Context, err error) interface{} {
				return &APIError{
					ErrorCode: "CUSTOM_ERROR",
					Details:   "Custom handled: " + err.Error(),
				}
			}

			options := []CallOption{
				ReplaceCallErrorModel(customErrorModel),
				WithErrorHandler(customErrorHandler),
			}
			options = append(options, tt.options...)

			// Configure operation
			doc.Post("/test").Call(tt.controller, options...)

			require.NoError(t, arrestDoc.Err())

			// The key test here is that the operation configures correctly
			// The actual runtime behavior would need integration tests
		})
	}
}
