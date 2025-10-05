package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/zostay/arrest-go"

	arrestgin "github.com/zostay/arrest-go/gin"
)

// Pet represents a pet in our system
type Pet struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Tag  string `json:"tag"`
}

// Pets is an array of pets
type Pets []Pet

// Error represents an error response
type Error struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

// Error implements the error interface for Error.
func (e *Error) Error() string {
	return e.Message
}

// CreatePetRequest represents the input for creating a pet
type CreatePetRequest struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Tag  string `json:"tag"`
	Type string `json:"type" openapi:",in=query"`
}

// GetPetRequest represents the input for getting a pet by ID
type GetPetRequest struct {
	// PetId is the ID of the pet to retrieve
	PetId string `json:"petId" openapi:",in=path"`
}

// PetListRequest represents the input for listing pets
type PetListRequest struct {
	Type string `json:"type" openapi:",in=query"`
	// Limit specifies the maximum number of pets to return
	Limit int32 `json:"limit" openapi:",in=query,required"`
}

// Simple in-memory storage for demonstration
var pets = []Pet{
	{ID: 1, Name: "Fluffy", Tag: "cat"},
	{ID: 2, Name: "Buddy", Tag: "dog"},
	{ID: 3, Name: "Charlie", Tag: "dog"},
}
var nextID int64 = 4

// Controller functions
func CreatePet(ctx context.Context, req CreatePetRequest) (*Pet, error) {
	pet := Pet{
		ID:   nextID,
		Name: req.Name,
		Tag:  req.Tag,
	}
	nextID++
	pets = append(pets, pet)
	return &pet, nil
}

func GetPet(ctx context.Context, req GetPetRequest) (*Pet, error) {
	for _, pet := range pets {
		if pet.ID == parseID(req.PetId) {
			return &pet, nil
		}
	}
	return nil, &Error{
		Code:    404,
		Message: "Pet not found",
	}
}

func ListPets(ctx context.Context, req PetListRequest) (Pets, error) {
	result := pets

	// Apply limit
	if req.Limit > 0 && int(req.Limit) < len(result) {
		result = result[:req.Limit]
	}

	return result, nil
}

// Helper function to parse string ID to int64
func parseID(id string) int64 {
	// In a real implementation, you'd use strconv.ParseInt
	// For simplicity, just return 1 for any string
	return 1
}

func BuildDoc(router gin.IRoutes) (*arrestgin.Document, error) {
	// Create arrest document
	arrestDoc, err := arrest.NewDocument("Pet Store API with Call Method Example")
	if err != nil {
		log.Fatal("Failed to create arrest document:", err)
	}

	// Set up info to match expected output
	arrestDoc.Version("1.0.0")
	arrestDoc.Description("Demonstration of the Call method that automatically generates handlers")
	arrestDoc.AddServer("http://petstore.swagger.io/v1")

	// Create gin router and document
	doc := arrestgin.NewDocument(arrestDoc, router)

	// Define API operations using the new Call method
	// Notice how much simpler this is compared to manually writing handlers!

	doc.Get("/pets").
		OperationID("listPets").
		Tags("pets").
		Summary("List all pets").
		Call(ListPets).
		Response("200", func(r *arrest.Response) {
			r.Description("PetListResponse represents the response for listing pets").
				Content("application/json", arrest.ModelFrom[Pets]())
		})

	doc.Post("/pets").
		OperationID("createPets").
		Tags("pets").
		Summary("Create a pet").
		Call(CreatePet).
		Response("201", func(r *arrest.Response) {
			r.Description("Null response")
		})

	doc.Get("/pets/{petId}").
		OperationID("showByPetId").
		Tags("pets").
		Summary("Info for a specific pet").
		Call(GetPet).
		Response("200", func(r *arrest.Response) {
			r.Description("Expected response to a valid request").
				Content("application/json", arrest.ModelFrom[Pet]())
		})

	// Check for any errors in the document setup
	if err := arrestDoc.Err(); err != nil {
		return nil, err
	}

	return doc, nil
}

// BuildDocString creates the OpenAPI document using Call method and returns it as a string
func BuildDocString() string {
	gin.SetMode(gin.ReleaseMode) // Suppress gin debug messages
	router := gin.New()
	doc, err := BuildDoc(router)
	if err != nil {
		panic(err)
	}

	rend, err := doc.OpenAPI.Render()
	if err != nil {
		panic(err)
	}

	return string(rend)
}

func main() {
	// Check if we're in test mode
	if len(os.Args) > 1 && os.Args[1] == "test" {
		fmt.Print(BuildDocString())
		return
	}

	// Create gin router and document
	router := gin.Default()
	doc, err := BuildDoc(router)
	if err != nil {
		log.Fatal("Failed to build document:", err)
	}

	// Add a route to serve the OpenAPI spec
	router.GET("/openapi.yaml", func(c *gin.Context) {
		openAPI, err := doc.OpenAPI.Render()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Header("Content-Type", "application/yaml")
		c.String(http.StatusOK, string(openAPI))
	})

	// Add a simple frontend to test the API
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Pet Store API Demo",
		})
	})

	// Load HTML template
	router.LoadHTMLGlob("*.html")

	log.Println("🚀 Server starting on :8080")
	log.Println("📖 OpenAPI spec available at: http://localhost:8080/openapi.yaml")
	log.Println("🌐 Demo UI available at: http://localhost:8080")
	log.Println()
	log.Println("Try these API endpoints:")
	log.Println("  POST   /pets              - Create a pet")
	log.Println("  GET    /pets              - List pets")
	log.Println("  GET    /pets/{id}         - Get pet by ID")
	log.Println("  PUT    /pets/{id}         - Update pet")
	log.Println()

	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
