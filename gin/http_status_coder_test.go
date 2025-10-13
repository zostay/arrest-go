package gin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	ginHTTP "github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zostay/arrest-go"
)

// CustomErrorResponse represents an error with a custom HTTP status code
type CustomErrorResponse struct {
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
}

func (e *CustomErrorResponse) Error() string {
	return e.Message
}

func (e *CustomErrorResponse) HTTPStatusCode() int {
	return e.StatusCode
}

// CreatedResponse represents a response with a custom HTTP status code
type CreatedResponse struct {
	ID      int    `json:"id"`
	Message string `json:"message"`
}

func (r *CreatedResponse) HTTPStatusCode() int {
	return http.StatusCreated
}

// Test controller functions
func controllerWithCustomError(ctx context.Context, input struct{}) (*struct{}, error) {
	return nil, &CustomErrorResponse{
		Message:    "Resource not found",
		StatusCode: http.StatusNotFound,
	}
}

func controllerWithCreatedResponse(ctx context.Context, input struct{}) (*CreatedResponse, error) {
	return &CreatedResponse{
		ID:      123,
		Message: "Resource created successfully",
	}, nil
}

func TestHTTPStatusCoder_CustomErrorStatusCode(t *testing.T) {
	t.Parallel()

	// Setup
	doc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	router := ginHTTP.New()
	ginDoc := NewDocument(doc, router)

	// Configure operation with controller that returns custom error
	ginDoc.Get("/test").Call(controllerWithCustomError)

	require.NoError(t, doc.Err())

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Verify custom status code is used
	assert.Equal(t, http.StatusNotFound, w.Code)

	// Verify response body
	var response CustomErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Resource not found", response.Message)
}

func TestHTTPStatusCoder_CustomResponseStatusCode(t *testing.T) {
	t.Parallel()

	// Setup
	doc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	router := ginHTTP.New()
	ginDoc := NewDocument(doc, router)

	// Configure operation with controller that returns custom response
	ginDoc.Post("/test").Call(controllerWithCreatedResponse)

	require.NoError(t, doc.Err())

	// Create request
	body := bytes.NewBufferString("{}")
	req := httptest.NewRequest(http.MethodPost, "/test", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Verify custom status code is used
	assert.Equal(t, http.StatusCreated, w.Code)

	// Verify response body
	var response CreatedResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, 123, response.ID)
	assert.Equal(t, "Resource created successfully", response.Message)
}

func TestHTTPStatusCoder_WithCustomErrorHandler(t *testing.T) {
	t.Parallel()

	// Setup
	doc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	router := ginHTTP.New()
	ginDoc := NewDocument(doc, router)

	// Custom error handler that returns a custom error response
	customErrorHandler := func(ctx *ginHTTP.Context, err error) interface{} {
		return &CustomErrorResponse{
			Message:    "Custom handled: " + err.Error(),
			StatusCode: http.StatusBadRequest,
		}
	}

	// Configure operation with custom error handler
	ginDoc.Get("/test").Call(controllerWithCustomError, WithErrorHandler(customErrorHandler))

	require.NoError(t, doc.Err())

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Verify custom error handler's status code is used
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Verify response body from custom handler
	var response CustomErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Custom handled: Resource not found", response.Message)
}
