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
	//nolint:staticcheck // this is example code
	return Pet{
		ID:   req.ID,
		Name: req.Name,
		Type: req.Type,
	}, nil
}

func init() {
	gin.SetMode(gin.TestMode)
}

func TestCallMethod_ValidController(t *testing.T) {
	t.Parallel()

	// Create arrest document and gin document
	arrestDoc, err := arrest.NewDocument("test")
	require.NoError(t, err)

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

	router := gin.New()
	doc := NewDocument(arrestDoc, router)

	// Test with options
	type CustomError struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	customErrorModel := arrest.ModelFrom[CustomError](arrestDoc)

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

func TestCallMethod_WithRequestComponent(t *testing.T) {
	t.Parallel()

	arrestDoc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	router := gin.New()
	doc := NewDocument(arrestDoc, router)

	// Test with request component registration
	doc.Post("/pets").Call(CreatePet, WithRequestComponent())

	assert.NoError(t, arrestDoc.Err())

	// Verify OpenAPI spec includes request component
	openAPI, err := arrestDoc.OpenAPI.Render()
	require.NoError(t, err)

	spec := string(openAPI)
	assert.Contains(t, spec, "components:")
	assert.Contains(t, spec, "schemas:")
	assert.Contains(t, spec, "github.com.zostay.arrest-go.gin.CreatePetRequest:")
	assert.Contains(t, spec, "$ref: '#/components/schemas/github.com.zostay.arrest-go.gin.CreatePetRequest'")
}

func TestCallMethod_WithResponseComponent(t *testing.T) {
	t.Parallel()

	arrestDoc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	router := gin.New()
	doc := NewDocument(arrestDoc, router)

	// Test with response component registration
	doc.Post("/pets").Call(CreatePet, WithResponseComponent())

	assert.NoError(t, arrestDoc.Err())

	// Verify OpenAPI spec includes response component
	openAPI, err := arrestDoc.OpenAPI.Render()
	require.NoError(t, err)

	spec := string(openAPI)
	assert.Contains(t, spec, "components:")
	assert.Contains(t, spec, "schemas:")
	assert.Contains(t, spec, "github.com.zostay.arrest-go.gin.Pet:")
	assert.Contains(t, spec, "$ref: '#/components/schemas/github.com.zostay.arrest-go.gin.Pet'")
}

func TestCallMethod_WithComponents(t *testing.T) {
	t.Parallel()

	arrestDoc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	router := gin.New()
	doc := NewDocument(arrestDoc, router)

	// Test with both request and response component registration
	doc.Post("/pets").Call(CreatePet, WithComponents())

	assert.NoError(t, arrestDoc.Err())

	// Verify OpenAPI spec includes both components
	openAPI, err := arrestDoc.OpenAPI.Render()
	require.NoError(t, err)

	spec := string(openAPI)
	assert.Contains(t, spec, "components:")
	assert.Contains(t, spec, "schemas:")
	assert.Contains(t, spec, "github.com.zostay.arrest-go.gin.CreatePetRequest:")
	assert.Contains(t, spec, "github.com.zostay.arrest-go.gin.Pet:")
	assert.Contains(t, spec, "$ref: '#/components/schemas/github.com.zostay.arrest-go.gin.CreatePetRequest'")
	assert.Contains(t, spec, "$ref: '#/components/schemas/github.com.zostay.arrest-go.gin.Pet'")
}

func TestCallMethod_WithoutComponents(t *testing.T) {
	t.Parallel()

	arrestDoc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	router := gin.New()
	doc := NewDocument(arrestDoc, router)

	// Test without component registration (default behavior)
	doc.Post("/pets").Call(CreatePet)

	assert.NoError(t, arrestDoc.Err())

	// Verify OpenAPI spec does not include components (inline schemas)
	openAPI, err := arrestDoc.OpenAPI.Render()
	require.NoError(t, err)

	spec := string(openAPI)
	// Should not have component references, but should have inline schemas
	assert.NotContains(t, spec, "$ref: '#/components/schemas/github.com.zostay.arrest-go.gin.CreatePetRequest'")
	assert.NotContains(t, spec, "$ref: '#/components/schemas/github.com.zostay.arrest-go.gin.Pet'")
	// But should still have the schema properties inline
	assert.Contains(t, spec, "name:")
	assert.Contains(t, spec, "type: string")
}

// Test polymorphic types for requests and responses
type PolymorphicPetRequest struct {
	PetType string            `json:"petType" openapi:",discriminator,defaultMapping=dog"`
	Dog     PolymorphicDog    `json:",inline,omitempty" openapi:",oneOf,mapping=dog"`
	Cat     PolymorphicCat    `json:",inline,omitempty" openapi:",oneOf,mapping=cat"`
	Bird    PolymorphicBird   `json:",inline,omitempty" openapi:",oneOf,mapping=bird"`
}

type PolymorphicDog struct {
	Breed string `json:"breed"`
	Name  string `json:"name"`
}

type PolymorphicCat struct {
	Lives int    `json:"lives"`
	Name  string `json:"name"`
}

type PolymorphicBird struct {
	CanFly bool   `json:"canFly"`
	Name   string `json:"name"`
}

// Polymorphic response type
type PolymorphicPetResponse struct {
	ID      int               `json:"id"`
	PetType string            `json:"petType" openapi:",discriminator,defaultMapping=dog"`
	Dog     PolymorphicDog    `json:",inline,omitempty" openapi:",oneOf,mapping=dog"`
	Cat     PolymorphicCat    `json:",inline,omitempty" openapi:",oneOf,mapping=cat"`
	Bird    PolymorphicBird   `json:",inline,omitempty" openapi:",oneOf,mapping=bird"`
}

// Polymorphic error response type
type PolymorphicError struct {
	ErrorType string            `json:"errorType" openapi:",discriminator,defaultMapping=validation"`
	Validation ValidationError  `json:",inline,omitempty" openapi:",oneOf,mapping=validation"`
	Internal   InternalError    `json:",inline,omitempty" openapi:",oneOf,mapping=internal"`
	NotFound   NotFoundError    `json:",inline,omitempty" openapi:",oneOf,mapping=not_found"`
}

type ValidationError struct {
	Status  string            `json:"status"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

type InternalError struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	RequestID string `json:"requestId,omitempty"`
}

type NotFoundError struct {
	Status   string `json:"status"`
	Message  string `json:"message"`
	Resource string `json:"resource"`
}

// Controller functions using polymorphic types
func CreatePolymorphicPet(ctx context.Context, req PolymorphicPetRequest) (PolymorphicPetResponse, error) {
	response := PolymorphicPetResponse{
		ID:      1,
		PetType: req.PetType,
	}

	// Copy the polymorphic data
	if req.PetType == "dog" {
		response.Dog = req.Dog
	} else if req.PetType == "cat" {
		response.Cat = req.Cat
	} else if req.PetType == "bird" {
		response.Bird = req.Bird
	}

	return response, nil
}

func GetPolymorphicPet(ctx context.Context, req GetPetRequest) (PolymorphicPetResponse, error) {
	return PolymorphicPetResponse{
		ID:      req.ID,
		PetType: "cat",
		Cat: PolymorphicCat{
			Lives: 9,
			Name:  "Fluffy",
		},
	}, nil
}

func TestCallMethod_PolymorphicRequest(t *testing.T) {
	t.Parallel()

	arrestDoc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	router := gin.New()
	doc := NewDocument(arrestDoc, router)

	// Test the Call method with polymorphic request
	doc.Post("/polymorphic-pets").Call(CreatePolymorphicPet)

	// Verify no errors
	assert.NoError(t, arrestDoc.Err())

	// Verify OpenAPI spec was generated correctly
	openAPI, err := arrestDoc.OpenAPI.Render()
	require.NoError(t, err)

	spec := string(openAPI)
	assert.Contains(t, spec, "/polymorphic-pets")
	assert.Contains(t, spec, "post:")
	assert.Contains(t, spec, "requestBody:")
	assert.Contains(t, spec, "oneOf:")
	assert.Contains(t, spec, "discriminator:")
	assert.Contains(t, spec, "propertyName: petType")
	assert.Contains(t, spec, "defaultMapping: dog")
	assert.Contains(t, spec, "breed:")      // Dog properties
	assert.Contains(t, spec, "lives:")      // Cat properties
	assert.Contains(t, spec, "canFly:")     // Bird properties
}

func TestCallMethod_PolymorphicResponse(t *testing.T) {
	t.Parallel()

	arrestDoc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	router := gin.New()
	doc := NewDocument(arrestDoc, router)

	// Test the Call method with polymorphic response
	doc.Get("/polymorphic-pets/{id}").Call(GetPolymorphicPet)

	// Verify no errors
	assert.NoError(t, arrestDoc.Err())

	// Verify OpenAPI spec was generated correctly
	openAPI, err := arrestDoc.OpenAPI.Render()
	require.NoError(t, err)

	spec := string(openAPI)
	assert.Contains(t, spec, "/polymorphic-pets/{id}")
	assert.Contains(t, spec, "get:")
	assert.Contains(t, spec, "responses:")
	assert.Contains(t, spec, "\"200\":")     // Quoted key in YAML
	assert.Contains(t, spec, "oneOf:")
	assert.Contains(t, spec, "discriminator:")
	assert.Contains(t, spec, "propertyName: petType")
	// Should contain polymorphic properties in oneOf
	assert.Contains(t, spec, "breed:")      // Dog properties
	assert.Contains(t, spec, "lives:")      // Cat properties
	assert.Contains(t, spec, "canFly:")     // Bird properties
}

func TestCallMethod_PolymorphicWithComponents(t *testing.T) {
	t.Parallel()

	arrestDoc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	router := gin.New()
	doc := NewDocument(arrestDoc, router)

	// Test polymorphic types with component registration
	doc.Post("/polymorphic-pets").Call(CreatePolymorphicPet, WithComponents())

	assert.NoError(t, arrestDoc.Err())

	// Verify OpenAPI spec includes polymorphic components
	openAPI, err := arrestDoc.OpenAPI.Render()
	require.NoError(t, err)

	spec := string(openAPI)
	assert.Contains(t, spec, "components:")
	assert.Contains(t, spec, "schemas:")
	// Should contain component references for polymorphic types
	assert.Contains(t, spec, "PolymorphicPetRequest")
	assert.Contains(t, spec, "PolymorphicPetResponse")
}

func TestGeneratedHandler_PolymorphicRequest(t *testing.T) {
	t.Parallel()

	arrestDoc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	router := gin.New()
	doc := NewDocument(arrestDoc, router)

	// Register the handler
	doc.Post("/polymorphic-pets").Call(CreatePolymorphicPet)
	assert.NoError(t, arrestDoc.Err())

	// Create test request with polymorphic data (Dog)
	// Note: Go's JSON unmarshaling doesn't automatically handle discriminated unions
	// so we need to structure the request to match the actual Go struct layout
	reqBody := map[string]interface{}{
		"petType": "dog",
		"Dog": map[string]interface{}{
			"breed": "Golden Retriever",
			"name":  "Buddy",
		},
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/polymorphic-pets", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(resp, req)

	// Check response
	assert.Equal(t, http.StatusOK, resp.Code)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body.Bytes(), &result)
	require.NoError(t, err)


	assert.Equal(t, float64(1), result["id"])       // JSON unmarshals numbers as float64
	assert.Equal(t, "dog", result["petType"])

	// Access the polymorphic data from the nested Dog object
	dog, ok := result["Dog"].(map[string]interface{})
	require.True(t, ok, "Dog field should be a nested object")
	assert.Equal(t, "Golden Retriever", dog["breed"])
	assert.Equal(t, "Buddy", dog["name"])
}

func TestGeneratedHandler_PolymorphicRequestCat(t *testing.T) {
	t.Parallel()

	arrestDoc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	router := gin.New()
	doc := NewDocument(arrestDoc, router)

	// Register the handler
	doc.Post("/polymorphic-pets").Call(CreatePolymorphicPet)
	assert.NoError(t, arrestDoc.Err())

	// Create test request with polymorphic data (Cat)
	// Note: Go's JSON unmarshaling doesn't automatically handle discriminated unions
	// so we need to structure the request to match the actual Go struct layout
	reqBody := map[string]interface{}{
		"petType": "cat",
		"Cat": map[string]interface{}{
			"lives": 9,
			"name":  "Whiskers",
		},
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/polymorphic-pets", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(resp, req)

	// Check response
	assert.Equal(t, http.StatusOK, resp.Code)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body.Bytes(), &result)
	require.NoError(t, err)


	assert.Equal(t, float64(1), result["id"])       // JSON unmarshals numbers as float64
	assert.Equal(t, "cat", result["petType"])

	// Access the polymorphic data from the nested Cat object
	cat, ok := result["Cat"].(map[string]interface{})
	require.True(t, ok, "Cat field should be a nested object")
	assert.Equal(t, float64(9), cat["lives"])       // JSON unmarshals numbers as float64
	assert.Equal(t, "Whiskers", cat["name"])
}

func TestCallMethod_PolymorphicError(t *testing.T) {
	t.Parallel()

	arrestDoc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	router := gin.New()
	doc := NewDocument(arrestDoc, router)

	// Create polymorphic error models
	validationErrorModel := arrest.ModelFrom[ValidationError](arrestDoc)
	internalErrorModel := arrest.ModelFrom[InternalError](arrestDoc)
	notFoundErrorModel := arrest.ModelFrom[NotFoundError](arrestDoc)

	// Test polymorphic error models
	doc.Post("/polymorphic-pets").Call(CreatePolymorphicPet, WithPolymorphicError(
		validationErrorModel,
		internalErrorModel,
		notFoundErrorModel,
	))

	assert.NoError(t, arrestDoc.Err())

	// Verify OpenAPI spec includes polymorphic error response
	openAPI, err := arrestDoc.OpenAPI.Render()
	require.NoError(t, err)

	spec := string(openAPI)
	assert.Contains(t, spec, "default:")
	assert.Contains(t, spec, "oneOf:")
	// Should contain all error model properties
	assert.Contains(t, spec, "status:")     // Common to all error types
	assert.Contains(t, spec, "message:")    // Common to all error types
	assert.Contains(t, spec, "fields:")     // ValidationError specific
	assert.Contains(t, spec, "requestId:")  // InternalError specific
	assert.Contains(t, spec, "resource:")   // NotFoundError specific
}

func TestCallMethod_PolymorphicErrorWithDiscriminator(t *testing.T) {
	t.Parallel()

	arrestDoc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	router := gin.New()
	doc := NewDocument(arrestDoc, router)

	// Create polymorphic error model using struct tags
	polymorphicErrorModel := arrest.ModelFrom[PolymorphicError](arrestDoc)

	// Test with struct-based polymorphic error
	doc.Post("/polymorphic-pets").Call(CreatePolymorphicPet, WithCallErrorModel(polymorphicErrorModel))

	assert.NoError(t, arrestDoc.Err())

	// Verify OpenAPI spec includes discriminated error response
	openAPI, err := arrestDoc.OpenAPI.Render()
	require.NoError(t, err)

	spec := string(openAPI)
	assert.Contains(t, spec, "default:")
	assert.Contains(t, spec, "oneOf:")
	assert.Contains(t, spec, "discriminator:")
	assert.Contains(t, spec, "propertyName: errorType")
	assert.Contains(t, spec, "defaultMapping: validation")
	// Should contain all error type properties
	assert.Contains(t, spec, "status:")     // Common to error types
	assert.Contains(t, spec, "message:")    // Common to error types
	assert.Contains(t, spec, "fields:")     // ValidationError specific
	assert.Contains(t, spec, "requestId:")  // InternalError specific
	assert.Contains(t, spec, "resource:")   // NotFoundError specific
}
