package gin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zostay/arrest-go"
)

// Test types for controller functions
type CreatePetRequest struct {
	Name string `json:"name"`
	Type string `json:"type" openapi:",in=query"`
}

type Pet struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type GetPetRequest struct {
	ID int `json:"id" openapi:",in=path"`
}

type UpdatePetRequest struct {
	ID   int    `json:"id" openapi:",in=path"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// Controller functions for testing
func CreatePet(ctx context.Context, req CreatePetRequest) (Pet, error) {
	return Pet{
		ID:   1,
		Name: req.Name,
		Type: req.Type,
	}, nil
}

func GetPet(ctx context.Context, req GetPetRequest) (Pet, error) {
	return Pet{
		ID:   req.ID,
		Name: "Fluffy",
		Type: "cat",
	}, nil
}

func UpdatePet(ctx context.Context, req UpdatePetRequest) (Pet, error) {
	return Pet{
		ID:   req.ID,
		Name: req.Name,
		Type: req.Type,
	}, nil
}

func TestCallMethod_ValidController(t *testing.T) {
	t.Parallel()

	// Create arrest document and gin document
	arrestDoc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	doc := NewDocument(arrestDoc, router)

	// Test the Call method
	doc.Post("/pets").Call(CreatePet)

	// Verify no errors
	assert.NoError(t, arrestDoc.Err())

	// Verify OpenAPI spec was generated correctly
	openAPI, err := arrestDoc.OpenAPI.Render()
	require.NoError(t, err)

	spec := string(openAPI)
	assert.Contains(t, spec, "/pets")
	assert.Contains(t, spec, "post:")
	assert.Contains(t, spec, "requestBody:")
	assert.Contains(t, spec, "application/json")
	assert.Contains(t, spec, "responses:")
	assert.Contains(t, spec, "default:")
}

func TestCallMethod_GetRequest(t *testing.T) {
	t.Parallel()

	arrestDoc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	doc := NewDocument(arrestDoc, router)

	// Test GET request (should not have request body)
	doc.Get("/pets/{id}").Call(GetPet)

	assert.NoError(t, arrestDoc.Err())

	openAPI, err := arrestDoc.OpenAPI.Render()
	require.NoError(t, err)

	spec := string(openAPI)
	assert.Contains(t, spec, "/pets/{id}")
	assert.Contains(t, spec, "get:")
	// GET requests should not have requestBody
	assert.NotContains(t, spec, "requestBody:")
	assert.Contains(t, spec, "responses:")
}

func TestCallMethod_WithOptions(t *testing.T) {
	t.Parallel()

	arrestDoc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	doc := NewDocument(arrestDoc, router)

	// Test with options
	type CustomError struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	customErrorModel := arrest.ModelFrom[CustomError]()

	doc.Post("/pets").Call(CreatePet,
		WithCallErrorModel(customErrorModel),
		WithPanicProtection(),
	)

	assert.NoError(t, arrestDoc.Err())
}

func TestCallMethod_InvalidController(t *testing.T) {
	t.Parallel()

	arrestDoc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	doc := NewDocument(arrestDoc, router)

	// Test with invalid controller (not a function)
	operation := doc.Post("/pets").Call("not a function")

	// Should have error - check the operation's error since that's where withErr adds it
	assert.Error(t, operation.Err())
}

func TestCallMethod_WrongSignature(t *testing.T) {
	t.Parallel()

	arrestDoc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	doc := NewDocument(arrestDoc, router)

	// Test with wrong signature
	wrongFunc := func(s string) string { return s }
	operation := doc.Post("/pets").Call(wrongFunc)

	// Should have error - check the operation's error since that's where withErr adds it
	assert.Error(t, operation.Err())
}

func TestGeneratedHandler_POST(t *testing.T) {
	t.Parallel()

	arrestDoc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	doc := NewDocument(arrestDoc, router)

	// Register the handler
	doc.Post("/pets").Call(CreatePet)
	assert.NoError(t, arrestDoc.Err())

	// Create test request
	reqBody := CreatePetRequest{
		Name: "Buddy",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/pets?type=dog", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(resp, req)

	// Check response
	assert.Equal(t, http.StatusOK, resp.Code)

	var result Pet
	err = json.Unmarshal(resp.Body.Bytes(), &result)
	require.NoError(t, err)

	assert.Equal(t, "Buddy", result.Name)
	assert.Equal(t, "dog", result.Type)
	assert.Equal(t, 1, result.ID)
}

func TestGeneratedHandler_GET(t *testing.T) {
	t.Parallel()

	arrestDoc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	doc := NewDocument(arrestDoc, router)

	// Register the handler
	doc.Get("/pets/:id").Call(GetPet)
	assert.NoError(t, arrestDoc.Err())

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/pets/123", nil)
	resp := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(resp, req)

	// Check response
	assert.Equal(t, http.StatusOK, resp.Code)

	var result Pet
	err = json.Unmarshal(resp.Body.Bytes(), &result)
	require.NoError(t, err)

	assert.Equal(t, 123, result.ID)
	assert.Equal(t, "Fluffy", result.Name)
	assert.Equal(t, "cat", result.Type)
}

func TestGeneratedHandler_PUT(t *testing.T) {
	t.Parallel()

	arrestDoc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	doc := NewDocument(arrestDoc, router)

	// Register the handler
	doc.Put("/pets/:id").Call(UpdatePet)
	assert.NoError(t, arrestDoc.Err())

	// Create test request with mixed path param and body
	reqBody := map[string]interface{}{
		"name": "Updated Name",
		"type": "updated type",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/pets/456", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(resp, req)

	// Check response
	assert.Equal(t, http.StatusOK, resp.Code)

	var result Pet
	err = json.Unmarshal(resp.Body.Bytes(), &result)
	require.NoError(t, err)

	assert.Equal(t, 456, result.ID)
	assert.Equal(t, "Updated Name", result.Name)
	assert.Equal(t, "updated type", result.Type)
}

func TestGeneratedHandler_ValidationError(t *testing.T) {
	t.Parallel()

	arrestDoc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	doc := NewDocument(arrestDoc, router)

	// Register the handler
	doc.Get("/pets/:id").Call(GetPet)
	assert.NoError(t, arrestDoc.Err())

	// Create test request with invalid path parameter
	req := httptest.NewRequest(http.MethodGet, "/pets/not-a-number", nil)
	resp := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(resp, req)

	// Check response
	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var errResp ErrorResponse
	err = json.Unmarshal(resp.Body.Bytes(), &errResp)
	require.NoError(t, err)

	assert.Equal(t, "error", errResp.Status)
	assert.Equal(t, "validation", errResp.Type)
	assert.Contains(t, errResp.Message, "convert")
}

func TestGeneratedHandler_ControllerError(t *testing.T) {
	t.Parallel()

	// Controller that returns an error
	errorController := func(ctx context.Context, req GetPetRequest) (Pet, error) {
		return Pet{}, assert.AnError
	}

	arrestDoc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	doc := NewDocument(arrestDoc, router)

	// Register the handler
	doc.Get("/pets/:id").Call(errorController)
	assert.NoError(t, arrestDoc.Err())

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/pets/123", nil)
	resp := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(resp, req)

	// Check response
	assert.Equal(t, http.StatusInternalServerError, resp.Code)

	var errResp ErrorResponse
	err = json.Unmarshal(resp.Body.Bytes(), &errResp)
	require.NoError(t, err)

	assert.Equal(t, "error", errResp.Status)
	assert.Equal(t, "internal", errResp.Type)
	assert.Contains(t, errResp.Message, "assert.AnError")
}

func TestGeneratedHandler_PanicProtection(t *testing.T) {
	t.Parallel()

	// Controller that panics
	panicController := func(ctx context.Context, req GetPetRequest) (Pet, error) {
		panic("test panic")
	}

	arrestDoc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	doc := NewDocument(arrestDoc, router)

	// Register the handler with panic protection
	doc.Get("/pets/:id").Call(panicController, WithPanicProtection())
	assert.NoError(t, arrestDoc.Err())

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/pets/123", nil)
	resp := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(resp, req)

	// Check response
	assert.Equal(t, http.StatusInternalServerError, resp.Code)

	var errResp ErrorResponse
	err = json.Unmarshal(resp.Body.Bytes(), &errResp)
	require.NoError(t, err)

	assert.Equal(t, "error", errResp.Status)
	assert.Equal(t, "internal", errResp.Type)
	assert.Contains(t, errResp.Message, "test panic")
}

func TestValidateControllerSignature(t *testing.T) {
	t.Parallel()

	arrestDoc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	op := &Operation{
		Operation: *arrestDoc.Get("/test"),
		method:    "GET",
		pattern:   "/test",
		r:         router,
	}

	// Valid controller
	err = op.validateControllerSignature(reflect.TypeOf(GetPet))
	assert.NoError(t, err)

	// Not a function
	err = op.validateControllerSignature(reflect.TypeOf("string"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a function")

	// Wrong number of parameters
	wrongParams := func(s string) (Pet, error) { return Pet{}, nil }
	err = op.validateControllerSignature(reflect.TypeOf(wrongParams))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exactly 2 parameters")

	// Wrong number of return values
	wrongReturns := func(ctx context.Context, req GetPetRequest) Pet { return Pet{} }
	err = op.validateControllerSignature(reflect.TypeOf(wrongReturns))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exactly 2 return values")

	// First parameter not context
	noContext := func(s string, req GetPetRequest) (Pet, error) { return Pet{}, nil }
	err = op.validateControllerSignature(reflect.TypeOf(noContext))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context.Context")

	// Second return not error
	noError := func(ctx context.Context, req GetPetRequest) (Pet, string) { return Pet{}, "" }
	err = op.validateControllerSignature(reflect.TypeOf(noError))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "implement error")
}

func TestHasBodyFields(t *testing.T) {
	t.Parallel()

	// Struct with mixed fields
	assert.True(t, hasBodyFields(reflect.TypeOf(UpdatePetRequest{})))

	// Struct with only query/path fields
	type OnlyParams struct {
		ID   int    `openapi:",in=path"`
		Name string `openapi:",in=query"`
	}
	assert.False(t, hasBodyFields(reflect.TypeOf(OnlyParams{})))

	// Non-struct type
	assert.True(t, hasBodyFields(reflect.TypeOf("")))
}

func TestParameterGeneration(t *testing.T) {
	t.Parallel()

	arrestDoc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	doc := NewDocument(arrestDoc, router)

	// Test that parameters are generated for types with parameter tags
	doc.Get("/pets/{id}").Call(GetPet)
	doc.Post("/pets").Call(CreatePet) // Has query parameter

	assert.NoError(t, arrestDoc.Err())

	// Verify OpenAPI spec includes parameters
	openAPI, err := arrestDoc.OpenAPI.Render()
	require.NoError(t, err)

	spec := string(openAPI)

	// Check that path parameter is included for GET /pets/{id}
	assert.Contains(t, spec, "/pets/{id}")
	assert.Contains(t, spec, "parameters:")
	assert.Contains(t, spec, "name: id")
	assert.Contains(t, spec, "in: path")

	// Check that query parameter is included for POST /pets
	assert.Contains(t, spec, "name: type")
	assert.Contains(t, spec, "in: query")
}
