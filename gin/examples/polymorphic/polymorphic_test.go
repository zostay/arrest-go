package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zostay/arrest-go"
	arrestgin "github.com/zostay/arrest-go/gin"
)

// TestCreateAnimal_Dog tests creating a dog with polymorphic input
func TestCreateAnimal_Dog(t *testing.T) {
	doc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	ginDoc := arrestgin.NewDocument(doc, router)

	ginDoc.Post("/animals").Call(CreateAnimal)
	require.NoError(t, doc.Err())

	// Test data for a dog - flattened polymorphic structure
	reqBody := CreateAnimalRequest{
		AnimalType: "dog",
		Dog: Dog{
			Breed:  "Labrador",
			Name:   "Buddy",
			IsGood: true,
		},
		Source: "test",
	}
	jsonBody, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/animals?source=test", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	// Debug: print response body if there's an error
	if resp.Code != http.StatusOK {
		t.Logf("Response status: %d", resp.Code)
		t.Logf("Response body: %s", resp.Body.String())
	}

	assert.Equal(t, http.StatusOK, resp.Code)

	var result AnimalResponse
	err = json.Unmarshal(resp.Body.Bytes(), &result)
	require.NoError(t, err)

	assert.Equal(t, int64(12345), result.ID)
	assert.Equal(t, "dog", result.AnimalType)
	assert.Equal(t, "Labrador", result.Dog.Breed)
	assert.Equal(t, "Buddy", result.Dog.Name)
	assert.True(t, result.Dog.IsGood)
}

// TestCreateAnimal_Cat tests creating a cat with polymorphic input
func TestCreateAnimal_Cat(t *testing.T) {
	doc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	ginDoc := arrestgin.NewDocument(doc, router)

	ginDoc.Post("/animals").Call(CreateAnimal)
	require.NoError(t, doc.Err())

	// Test data for a cat - flattened polymorphic structure
	reqBody := CreateAnimalRequest{
		AnimalType: "cat",
		Cat: Cat{
			Lives: 7,
			Name:  "Whiskers",
			Color: "orange",
		},
		Source: "test",
	}
	jsonBody, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/animals?source=test", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var result AnimalResponse
	err = json.Unmarshal(resp.Body.Bytes(), &result)
	require.NoError(t, err)

	assert.Equal(t, int64(12345), result.ID)
	assert.Equal(t, "cat", result.AnimalType)
	assert.Equal(t, 7, result.Cat.Lives)
	assert.Equal(t, "Whiskers", result.Cat.Name)
	assert.Equal(t, "orange", result.Cat.Color)
}

// TestCreateAnimal_Bird tests creating a bird with polymorphic input
func TestCreateAnimal_Bird(t *testing.T) {
	doc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	ginDoc := arrestgin.NewDocument(doc, router)

	ginDoc.Post("/animals").Call(CreateAnimal)
	require.NoError(t, doc.Err())

	// Test data for a bird - Go's JSON unmarshaling requires nested structure for polymorphic types
	reqBody := CreateAnimalRequest{
		AnimalType: "bird",
		Bird: Bird{
				CanFly:   true,
				Name:     "Eagle",
				Species:  "Bald Eagle",
				Wingspan: 220,
		},
		Source: "test",
	}
	jsonBody, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/animals?source=test", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var result AnimalResponse
	err = json.Unmarshal(resp.Body.Bytes(), &result)
	require.NoError(t, err)

	assert.Equal(t, int64(12345), result.ID)
	assert.Equal(t, "bird", result.AnimalType)
	assert.True(t, result.Bird.CanFly)
	assert.Equal(t, "Eagle", result.Bird.Name)
	assert.Equal(t, "Bald Eagle", result.Bird.Species)
	assert.Equal(t, 220, result.Bird.Wingspan)
}

// TestCreateAnimal_ValidationError tests polymorphic error responses
func TestCreateAnimal_ValidationError(t *testing.T) {
	doc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	ginDoc := arrestgin.NewDocument(doc, router)

	// Create error models for polymorphic errors
	validationErrorModel := arrest.ModelFrom[ValidationError](doc, arrest.AsComponent())
	businessErrorModel := arrest.ModelFrom[BusinessError](doc, arrest.AsComponent())
	systemErrorModel := arrest.ModelFrom[SystemError](doc, arrest.AsComponent())

	ginDoc.Post("/animals").Call(CreateAnimal,
		arrestgin.WithPolymorphicError(validationErrorModel, businessErrorModel, systemErrorModel))
	require.NoError(t, doc.Err())

	// Test invalid cat with too many lives
	reqBody := CreateAnimalRequest{
		AnimalType: "cat",
		Cat: Cat{
				Lives: 10, // Invalid: more than 9 lives
				Name:  "Whiskers",
				Color: "orange",
		},
		Source: "test",
	}
	jsonBody, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/animals?source=test", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var errorResult ValidationError
	err = json.Unmarshal(resp.Body.Bytes(), &errorResult)
	require.NoError(t, err)

	assert.Equal(t, "lives", errorResult.Field)
	assert.Equal(t, "range", errorResult.Code)
	assert.Contains(t, errorResult.Message, "between 1 and 9")
}

// TestCreateAnimal_BusinessError tests business error responses
func TestCreateAnimal_BusinessError(t *testing.T) {
	doc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	ginDoc := arrestgin.NewDocument(doc, router)

	// Create error models for polymorphic errors
	validationErrorModel := arrest.ModelFrom[ValidationError](doc, arrest.AsComponent())
	businessErrorModel := arrest.ModelFrom[BusinessError](doc, arrest.AsComponent())
	systemErrorModel := arrest.ModelFrom[SystemError](doc, arrest.AsComponent())

	ginDoc.Post("/animals").Call(CreateAnimal,
		arrestgin.WithPolymorphicError(validationErrorModel, businessErrorModel, systemErrorModel))
	require.NoError(t, doc.Err())

	// Test business error with bad dog name
	reqBody := CreateAnimalRequest{
		AnimalType: "dog",
		Dog: Dog{
				Breed:  "Labrador",
				Name:   "BadDog", // This triggers a business error
				IsGood: true,
		},
		Source: "test",
	}
	jsonBody, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/animals?source=test", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var errorResult BusinessError
	err = json.Unmarshal(resp.Body.Bytes(), &errorResult)
	require.NoError(t, err)

	assert.Equal(t, "INVALID_NAME", errorResult.ErrorCode)
	assert.Contains(t, errorResult.Details, "BadDog")
	assert.Contains(t, errorResult.Action, "different name")
}

// TestCreateAnimal_SystemError tests system error responses
func TestCreateAnimal_SystemError(t *testing.T) {
	doc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	ginDoc := arrestgin.NewDocument(doc, router)

	// Create error models for polymorphic errors
	validationErrorModel := arrest.ModelFrom[ValidationError](doc, arrest.AsComponent())
	businessErrorModel := arrest.ModelFrom[BusinessError](doc, arrest.AsComponent())
	systemErrorModel := arrest.ModelFrom[SystemError](doc, arrest.AsComponent())

	ginDoc.Post("/animals").Call(CreateAnimal,
		arrestgin.WithPolymorphicError(validationErrorModel, businessErrorModel, systemErrorModel))
	require.NoError(t, doc.Err())

	// Test system error by setting source=error
	reqBody := CreateAnimalRequest{
		AnimalType: "dog",
		Dog: Dog{
				Breed:  "Labrador",
				Name:   "Buddy",
				IsGood: true,
		},
		Source: "error", // This triggers a system error
	}
	jsonBody, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/animals?source=error", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)

	var errorResult SystemError
	err = json.Unmarshal(resp.Body.Bytes(), &errorResult)
	require.NoError(t, err)

	assert.Equal(t, "database", errorResult.Component)
	assert.Equal(t, "high", errorResult.Severity)
	assert.Contains(t, errorResult.Message, "connection failed")
}

// TestCreateVehicle_Car tests creating a vehicle with component references
func TestCreateVehicle_Car(t *testing.T) {
	doc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	ginDoc := arrestgin.NewDocument(doc, router)

	ginDoc.Post("/vehicles").Call(CreateVehicle)
	require.NoError(t, doc.Err())

	// Test data for a car - flattened polymorphic structure
	reqBody := CreateVehicleRequest{
		VehicleType: "car",
		Car: &Car{
			Doors: 4,
			Brand: "Honda",
			Model: "Civic",
		},
		Owner: "testuser",
	}
	jsonBody, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/vehicles?owner=testuser", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var result VehicleResponse
	err = json.Unmarshal(resp.Body.Bytes(), &result)
	require.NoError(t, err)

	assert.Equal(t, int64(67890), result.ID)
	assert.Equal(t, "car", result.VehicleType)
	assert.Equal(t, "active", result.Status)
	require.NotNil(t, result.Car)
	assert.Equal(t, 4, result.Car.Doors)
	assert.Equal(t, "Honda", result.Car.Brand)
	assert.Equal(t, "Civic", result.Car.Model)
}

// TestCreateVehicle_ValidationError tests vehicle validation errors
func TestCreateVehicle_ValidationError(t *testing.T) {
	doc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	ginDoc := arrestgin.NewDocument(doc, router)

	validationErrorModel := arrest.ModelFrom[ValidationError](doc, arrest.AsComponent())
	ginDoc.Post("/vehicles").Call(CreateVehicle,
		arrestgin.WithPolymorphicError(validationErrorModel))
	require.NoError(t, doc.Err())

	// Test invalid car with too many doors
	reqBody := CreateVehicleRequest{
		VehicleType: "car",
		Car: &Car{
			Doors: 10, // Invalid: more than 5 doors
			Brand: "Honda",
			Model: "Civic",
		},
		Owner: "testuser",
	}
	jsonBody, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/vehicles?owner=testuser", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var errorResult ValidationError
	err = json.Unmarshal(resp.Body.Bytes(), &errorResult)
	require.NoError(t, err)

	assert.Equal(t, "doors", errorResult.Field)
	assert.Equal(t, "range", errorResult.Code)
	assert.Contains(t, errorResult.Message, "between 2 and 5")
}

// TestGetAnimal tests retrieving an animal by ID
func TestGetAnimal(t *testing.T) {
	doc, err := arrest.NewDocument("test")
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	ginDoc := arrestgin.NewDocument(doc, router)

	ginDoc.Get("/animals/{id}").Call(GetAnimal)
	require.NoError(t, doc.Err())

	req := httptest.NewRequest(http.MethodGet, "/animals/123", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var result AnimalResponse
	err = json.Unmarshal(resp.Body.Bytes(), &result)
	require.NoError(t, err)

	assert.Equal(t, int64(123), result.ID)
	assert.Equal(t, "dog", result.AnimalType)
	assert.Equal(t, "Golden Retriever", result.Dog.Breed)
	assert.Equal(t, "Max", result.Dog.Name)
	assert.True(t, result.Dog.IsGood)
}

// TestOpenAPISpecGeneration tests that the OpenAPI specification is generated correctly
func TestOpenAPISpecGeneration(t *testing.T) {
	doc, err := arrest.NewDocument("Polymorphic Test API")
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	ginDoc := arrestgin.NewDocument(doc, router)

	// Create error models for polymorphic errors
	validationErrorModel := arrest.ModelFrom[ValidationError](doc, arrest.AsComponent())
	businessErrorModel := arrest.ModelFrom[BusinessError](doc, arrest.AsComponent())
	systemErrorModel := arrest.ModelFrom[SystemError](doc, arrest.AsComponent())

	// Register endpoints with polymorphic support
	ginDoc.Post("/animals").Call(CreateAnimal,
		arrestgin.WithPolymorphicError(validationErrorModel, businessErrorModel, systemErrorModel),
		arrestgin.WithComponents())

	ginDoc.Get("/animals/{id}").Call(GetAnimal,
		arrestgin.WithPolymorphicError(validationErrorModel))

	ginDoc.Post("/vehicles").Call(CreateVehicle,
		arrestgin.WithPolymorphicError(validationErrorModel),
		arrestgin.WithComponents())

	require.NoError(t, doc.Err())

	// Generate OpenAPI spec
	openAPISpec, err := doc.OpenAPI.Render()
	require.NoError(t, err)

	spec := string(openAPISpec)

	// Verify polymorphic elements are present
	assert.Contains(t, spec, "oneOf:")
	assert.Contains(t, spec, "discriminator:")
	assert.Contains(t, spec, "propertyName: animalType")
	assert.Contains(t, spec, "propertyName: vehicleType")
	assert.Contains(t, spec, "defaultMapping: dog")
	assert.Contains(t, spec, "defaultMapping: car")

	// Verify polymorphic mappings
	assert.Contains(t, spec, "mapping:")
	assert.Contains(t, spec, "dog:")
	assert.Contains(t, spec, "cat:")
	assert.Contains(t, spec, "bird:")
	assert.Contains(t, spec, "car:")
	assert.Contains(t, spec, "truck:")
	assert.Contains(t, spec, "motorcycle:")

	// Verify component schemas are defined
	assert.Contains(t, spec, "components:")
	assert.Contains(t, spec, "schemas:")
	assert.Contains(t, spec, "ValidationError")
	assert.Contains(t, spec, "BusinessError")
	assert.Contains(t, spec, "SystemError")

	// Verify polymorphic properties are included
	assert.Contains(t, spec, "breed:")       // Dog property
	assert.Contains(t, spec, "lives:")       // Cat property
	assert.Contains(t, spec, "canFly:")      // Bird property
	assert.Contains(t, spec, "doors:")       // Car property
	assert.Contains(t, spec, "capacity:")    // Truck property
	assert.Contains(t, spec, "ccs:")         // Motorcycle property

	t.Logf("Generated OpenAPI spec:\n%s", spec)
}

// TestControllerFunctions_DirectCall tests the controller functions directly
func TestControllerFunctions_DirectCall(t *testing.T) {
	ctx := context.Background()

	// Test successful animal creation
	req := CreateAnimalRequest{
		AnimalType: "dog",
		Dog: Dog{
				Breed:  "Beagle",
				Name:   "Snoopy",
				IsGood: true,
		},
		Source: "test",
	}

	resp, err := CreateAnimal(ctx, req)
	require.NoError(t, err)

	assert.Equal(t, int64(12345), resp.ID)
	assert.Equal(t, "dog", resp.AnimalType)
	assert.Equal(t, "Beagle", resp.Dog.Breed)
	assert.Equal(t, "Snoopy", resp.Dog.Name)
	assert.True(t, resp.Dog.IsGood)

	// Test validation error
	invalidReq := CreateAnimalRequest{
		AnimalType: "cat",
		Cat: Cat{
				Lives: 15, // Invalid: more than 9 lives
				Name:  "Garfield",
				Color: "orange",
		},
		Source: "test",
	}

	_, err = CreateAnimal(ctx, invalidReq)
	require.Error(t, err)

	var validationErr ValidationError
	assert.ErrorAs(t, err, &validationErr)
	assert.Equal(t, "lives", validationErr.Field)
	assert.Equal(t, "range", validationErr.Code)
}