package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/zostay/arrest-go"
	arrestgin "github.com/zostay/arrest-go/gin"
)

// =============================================================================
// Polymorphic Animal Types
// =============================================================================

// Animal represents a polymorphic animal using implicit polymorphism with struct tags
type Animal struct {
	AnimalType string `json:"animalType" openapi:",discriminator,defaultMapping=dog"`
	Dog        Dog    `json:",inline,omitempty" openapi:",oneOf,mapping=dog"`
	Cat        Cat    `json:",inline,omitempty" openapi:",oneOf,mapping=cat"`
	Bird       Bird   `json:",inline,omitempty" openapi:",oneOf,mapping=bird"`
}

type Dog struct {
	Breed  string `json:"breed"`
	Name   string `json:"name"`
	IsGood bool   `json:"isGood"`
}

type Cat struct {
	Lives int    `json:"lives"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type Bird struct {
	CanFly   bool   `json:"canFly"`
	Name     string `json:"name"`
	Species  string `json:"species"`
	Wingspan int    `json:"wingspan"`
}

// =============================================================================
// Vehicle Types Using Component References
// =============================================================================

// Vehicle represents a polymorphic vehicle using component references
type Vehicle struct {
	VehicleType string      `json:"vehicleType" openapi:",discriminator,defaultMapping=car"`
	Car         *Car        `json:"car,omitempty" openapi:",oneOf,mapping=car,refName=Car"`
	Truck       *Truck      `json:"truck,omitempty" openapi:",oneOf,mapping=truck,refName=Truck"`
	Motorcycle  *Motorcycle `json:"motorcycle,omitempty" openapi:",oneOf,mapping=motorcycle,refName=Motorcycle"`
}

type Car struct {
	Doors int    `json:"doors"`
	Brand string `json:"brand"`
	Model string `json:"model"`
}

type Truck struct {
	Capacity int    `json:"capacity"`
	Brand    string `json:"brand"`
	AxleType string `json:"axleType"`
}

type Motorcycle struct {
	CCs   int    `json:"ccs"`
	Brand string `json:"brand"`
	Type  string `json:"type"`
}

// =============================================================================
// Error Types for Polymorphic Error Responses
// =============================================================================

// ValidationError represents validation errors
type ValidationError struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error on field %s: %s", e.Field, e.Message)
}

// BusinessError represents business logic errors
type BusinessError struct {
	ErrorCode string `json:"errorCode"`
	Details   string `json:"details"`
	Action    string `json:"suggestedAction"`
}

func (e BusinessError) Error() string {
	return fmt.Sprintf("business error %s: %s", e.ErrorCode, e.Details)
}

// SystemError represents system-level errors
type SystemError struct {
	Component string `json:"component"`
	Severity  string `json:"severity"`
	Message   string `json:"message"`
}

func (e SystemError) Error() string {
	return fmt.Sprintf("system error in %s: %s", e.Component, e.Message)
}

// =============================================================================
// Request/Response Types
// =============================================================================

// CreateAnimalRequest represents the input for creating an animal
type CreateAnimalRequest struct {
	AnimalType string `json:"animalType" openapi:",discriminator,defaultMapping=dog"`
	Dog        Dog    `json:",inline,omitempty" openapi:",oneOf,mapping=dog"`
	Cat        Cat    `json:",inline,omitempty" openapi:",oneOf,mapping=cat"`
	Bird       Bird   `json:",inline,omitempty" openapi:",oneOf,mapping=bird"`
	Source     string `json:"source" openapi:",in=query"`
}

// UpdateAnimalRequest represents the input for updating an animal
type UpdateAnimalRequest struct {
	ID         string `json:"id" openapi:",in=path"`
	AnimalType string `json:"animalType" openapi:",discriminator,defaultMapping=dog"`
	Dog        Dog    `json:",inline,omitempty" openapi:",oneOf,mapping=dog"`
	Cat        Cat    `json:",inline,omitempty" openapi:",oneOf,mapping=cat"`
	Bird       Bird   `json:",inline,omitempty" openapi:",oneOf,mapping=bird"`
}

// GetAnimalRequest represents the input for getting an animal by ID
type GetAnimalRequest struct {
	ID string `json:"id" openapi:",in=path"`
}

// AnimalResponse represents a single animal response
type AnimalResponse struct {
	ID        int64  `json:"id"`
	CreatedAt string `json:"createdAt"`
	Animal
}

// CreateVehicleRequest represents the input for creating a vehicle
type CreateVehicleRequest struct {
	VehicleType string      `json:"vehicleType" openapi:",discriminator,defaultMapping=car"`
	Car         *Car        `json:"car,omitempty" openapi:",oneOf,mapping=car,refName=Car"`
	Truck       *Truck      `json:"truck,omitempty" openapi:",oneOf,mapping=truck,refName=Truck"`
	Motorcycle  *Motorcycle `json:"motorcycle,omitempty" openapi:",oneOf,mapping=motorcycle,refName=Motorcycle"`
	Owner       string      `json:"owner" openapi:",in=query"`
}

// GetVehicleRequest represents the input for getting a vehicle by ID
type GetVehicleRequest struct {
	ID string `json:"id" openapi:",in=path"`
}

// VehicleResponse represents a single vehicle response
type VehicleResponse struct {
	ID        int64  `json:"id"`
	CreatedAt string `json:"createdAt"`
	Status    string `json:"status"`
	Vehicle
}

// =============================================================================
// Controller Functions
// =============================================================================

// CreateAnimal creates a new animal with polymorphic input
func CreateAnimal(ctx context.Context, req CreateAnimalRequest) (AnimalResponse, error) {
	// Validate the request based on animal type
	switch req.AnimalType {
	case "dog":
		if req.Dog.Name == "" {
			return AnimalResponse{}, ValidationError{
				Field:   "name",
				Code:    "required",
				Message: "Dog name is required",
			}
		}
		if req.Dog.Name == "BadDog" {
			return AnimalResponse{}, BusinessError{
				ErrorCode: "INVALID_NAME",
				Details:   "Dogs cannot be named 'BadDog'",
				Action:    "Choose a different name",
			}
		}
	case "cat":
		if req.Cat.Lives < 1 || req.Cat.Lives > 9 {
			return AnimalResponse{}, ValidationError{
				Field:   "lives",
				Code:    "range",
				Message: "Cats must have between 1 and 9 lives",
			}
		}
	case "bird":
		if req.Bird.Wingspan < 0 {
			return AnimalResponse{}, ValidationError{
				Field:   "wingspan",
				Code:    "min",
				Message: "Wingspan cannot be negative",
			}
		}
	default:
		return AnimalResponse{}, ValidationError{
			Field:   "animalType",
			Code:    "invalid",
			Message: "Animal type must be dog, cat, or bird",
		}
	}

	// Simulate system error for testing
	if req.Source == "error" {
		return AnimalResponse{}, SystemError{
			Component: "database",
			Severity:  "high",
			Message:   "Database connection failed",
		}
	}

	// Create successful response
	return AnimalResponse{
		ID:        12345,
		CreatedAt: "2023-10-13T10:30:00Z",
		Animal: Animal{
			AnimalType: req.AnimalType,
			Dog:        req.Dog,
			Cat:        req.Cat,
			Bird:       req.Bird,
		},
	}, nil
}

// GetAnimal retrieves an animal by ID
func GetAnimal(ctx context.Context, req GetAnimalRequest) (AnimalResponse, error) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		return AnimalResponse{}, ValidationError{
			Field:   "id",
			Code:    "invalid",
			Message: "ID must be a valid number",
		}
	}

	// Return a sample dog for demonstration
	return AnimalResponse{
		ID:        id,
		CreatedAt: "2023-10-13T10:00:00Z",
		Animal: Animal{
			AnimalType: "dog",
			Dog: Dog{
				Breed:  "Golden Retriever",
				Name:   "Max",
				IsGood: true,
			},
		},
	}, nil
}

// UpdateAnimal updates an existing animal
func UpdateAnimal(ctx context.Context, req UpdateAnimalRequest) (AnimalResponse, error) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		return AnimalResponse{}, ValidationError{
			Field:   "id",
			Code:    "invalid",
			Message: "ID must be a valid number",
		}
	}

	return AnimalResponse{
		ID:        id,
		CreatedAt: "2023-10-13T10:00:00Z",
		Animal: Animal{
			AnimalType: req.AnimalType,
			Dog:        req.Dog,
			Cat:        req.Cat,
			Bird:       req.Bird,
		},
	}, nil
}

// CreateVehicle creates a new vehicle with polymorphic input using component references
func CreateVehicle(ctx context.Context, req CreateVehicleRequest) (VehicleResponse, error) {
	// Validate the request based on vehicle type
	switch req.VehicleType {
	case "car":
		if req.Car == nil {
			return VehicleResponse{}, ValidationError{
				Field:   "car",
				Code:    "required",
				Message: "Car details are required when vehicleType is 'car'",
			}
		}
		if req.Car.Doors < 2 || req.Car.Doors > 5 {
			return VehicleResponse{}, ValidationError{
				Field:   "doors",
				Code:    "range",
				Message: "Car must have between 2 and 5 doors",
			}
		}
	case "truck":
		if req.Truck == nil {
			return VehicleResponse{}, ValidationError{
				Field:   "truck",
				Code:    "required",
				Message: "Truck details are required when vehicleType is 'truck'",
			}
		}
		if req.Truck.Capacity <= 0 {
			return VehicleResponse{}, ValidationError{
				Field:   "capacity",
				Code:    "min",
				Message: "Truck capacity must be positive",
			}
		}
	case "motorcycle":
		if req.Motorcycle == nil {
			return VehicleResponse{}, ValidationError{
				Field:   "motorcycle",
				Code:    "required",
				Message: "Motorcycle details are required when vehicleType is 'motorcycle'",
			}
		}
	default:
		return VehicleResponse{}, ValidationError{
			Field:   "vehicleType",
			Code:    "invalid",
			Message: "Vehicle type must be car, truck, or motorcycle",
		}
	}

	return VehicleResponse{
		ID:        67890,
		CreatedAt: "2023-10-13T11:00:00Z",
		Status:    "active",
		Vehicle: Vehicle{
			VehicleType: req.VehicleType,
			Car:         req.Car,
			Truck:       req.Truck,
			Motorcycle:  req.Motorcycle,
		},
	}, nil
}

// GetVehicle retrieves a vehicle by ID
func GetVehicle(ctx context.Context, req GetVehicleRequest) (VehicleResponse, error) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		return VehicleResponse{}, ValidationError{
			Field:   "id",
			Code:    "invalid",
			Message: "ID must be a valid number",
		}
	}

	// Return a sample car for demonstration
	return VehicleResponse{
		ID:        id,
		CreatedAt: "2023-10-13T10:30:00Z",
		Status:    "active",
		Vehicle: Vehicle{
			VehicleType: "car",
			Car: &Car{
				Doors: 4,
				Brand: "Toyota",
				Model: "Camry",
			},
		},
	}, nil
}

// =============================================================================
// Main Function
// =============================================================================

func main() {
	// Create OpenAPI document
	doc, err := arrest.NewDocument("Polymorphic API Example")
	if err != nil {
		log.Fatalf("Failed to create OpenAPI document: %v", err)
	}

	doc.Description("A comprehensive example demonstrating polymorphic types with the Gin Call method")
	doc.Version("1.0.0")
	doc.PackageMap("polymorphic.v1", "main")

	// Create Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// Create gin document wrapper
	ginDoc := arrestgin.NewDocument(doc, router)

	// Create error models for polymorphic error responses
	validationErrorModel := arrest.ModelFrom[ValidationError](doc, arrest.AsComponent()).
		Description("Validation error response")
	businessErrorModel := arrest.ModelFrom[BusinessError](doc, arrest.AsComponent()).
		Description("Business logic error response")
	systemErrorModel := arrest.ModelFrom[SystemError](doc, arrest.AsComponent()).
		Description("System error response")

	// Animal endpoints with polymorphic requests/responses and errors
	ginDoc.Post("/animals").
		Summary("Create a new animal").
		Description("Creates a new animal with polymorphic input supporting dogs, cats, and birds").
		Tags("animals").
		Call(CreateAnimal,
			arrestgin.WithPolymorphicError(validationErrorModel, businessErrorModel, systemErrorModel),
			arrestgin.WithResponseComponent(),
		)

	ginDoc.Get("/animals/{id}").
		Summary("Get an animal by ID").
		Description("Retrieves an animal by its unique identifier").
		Tags("animals").
		Call(GetAnimal,
			arrestgin.WithPolymorphicError(validationErrorModel),
		)

	ginDoc.Put("/animals/{id}").
		Summary("Update an animal").
		Description("Updates an existing animal with polymorphic input").
		Tags("animals").
		Call(UpdateAnimal,
			arrestgin.WithPolymorphicError(validationErrorModel),
			arrestgin.WithComponents(),
		)

	// Vehicle endpoints using component references
	ginDoc.Post("/vehicles").
		Summary("Create a new vehicle").
		Description("Creates a new vehicle with polymorphic input using component references").
		Tags("vehicles").
		Call(CreateVehicle,
			arrestgin.WithPolymorphicError(validationErrorModel),
			arrestgin.WithComponents(),
		)

	ginDoc.Get("/vehicles/{id}").
		Summary("Get a vehicle by ID").
		Description("Retrieves a vehicle by its unique identifier").
		Tags("vehicles").
		Call(GetVehicle,
			arrestgin.WithPolymorphicError(validationErrorModel),
		)

	// Check for any errors in document construction
	if err := doc.Err(); err != nil {
		log.Fatalf("OpenAPI document has errors: %v", err)
	}

	// Generate and save OpenAPI spec
	openAPISpec, err := doc.OpenAPI.Render()
	if err != nil {
		log.Fatalf("Failed to render OpenAPI spec: %v", err)
	}

	// Save OpenAPI specification
	if err := os.WriteFile("openapi.yaml", openAPISpec, 0644); err != nil {
		log.Printf("Warning: Failed to write OpenAPI spec to openapi.yaml: %v", err)
	} else {
		fmt.Println("OpenAPI specification written to openapi.yaml")
	}

	// Add a route to serve the OpenAPI spec
	router.GET("/openapi.yaml", func(c *gin.Context) {
		c.Header("Content-Type", "application/x-yaml")
		c.String(http.StatusOK, string(openAPISpec))
	})

	// Add a simple HTML page to view the API docs
	router.GET("/", func(c *gin.Context) {
		html := `<!DOCTYPE html>
<html>
<head>
    <title>Polymorphic API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@3.52.5/swagger-ui.css" />
    <style>
        html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
        *, *:before, *:after { box-sizing: inherit; }
        body { margin:0; background: #fafafa; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@3.52.5/swagger-ui-bundle.js"></script>
    <script>
        SwaggerUIBundle({
            url: '/openapi.yaml',
            dom_id: '#swagger-ui',
            deepLinking: true,
            presets: [
                SwaggerUIBundle.presets.apis,
                SwaggerUIBundle.presets.standalone
            ],
            plugins: [
                SwaggerUIBundle.plugins.DownloadUrl
            ],
            layout: "StandaloneLayout"
        });
    </script>
</body>
</html>`
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, html)
	})

	// Check if we should just generate the spec and exit (for CI/documentation)
	if len(os.Args) > 1 && os.Args[1] == "--generate-spec" {
		fmt.Println("OpenAPI specification generation complete")
		return
	}

	// Start server
	port := ":8080"
	fmt.Printf("Starting server on http://localhost%s\n", port)
	fmt.Printf("View API documentation at http://localhost%s\n", port)
	fmt.Printf("OpenAPI spec available at http://localhost%s/openapi.yaml\n", port)

	if err := router.Run(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
