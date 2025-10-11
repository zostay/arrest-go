package main

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/zostay/arrest-go"
	arrestGin "github.com/zostay/arrest-go/gin"
)

type CreatePetRequest struct {
	Pet Pet `json:"pet" openapi:",refName=Pet"`
}

type Pet struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Tag  string `json:"tag"`
}

type Pets []Pet

type Error struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

func handleListPets(c *gin.Context) {
	limitStr := c.Query("limit")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(400, Error{Code: 400, Message: "Invalid limit"})
		return
	}

	pets, err := ListPets(int32(limit))
	if err != nil {
		c.JSON(500, Error{Code: 500, Message: "Internal Server Error"})
		return
	}

	c.JSON(200, pets)
}

func ListPets(limit int32) (Pets, error) {
	// This is where you would put your implementation of ListPets
	return nil, nil
}

func handleCreatePets(c *gin.Context) {
	var pet Pet
	if err := c.ShouldBindJSON(&pet); err != nil {
		c.JSON(400, Error{Code: 400, Message: "Invalid input"})
		return
	}

	if err := CreatePets(pet); err != nil {
		c.JSON(500, Error{Code: 500, Message: "Internal Server Error"})
		return
	}

	c.JSON(201, nil)
}

func CreatePets(pet Pet) error {
	// This is where you would put your implementation of CreatePets
	return nil
}

func handleShowByPetID(c *gin.Context) {
	petIDStr := c.Param("petId")
	petID, err := strconv.Atoi(petIDStr)
	if err != nil {
		c.JSON(400, Error{Code: 400, Message: "Invalid pet ID"})
		return
	}

	pet, err := ShowByPetID(strconv.Itoa(petID))
	if err != nil {
		c.JSON(500, Error{Code: 500, Message: "Internal Server Error"})
		return
	}

	c.JSON(200, pet)
}

func ShowByPetID(petID string) (*Pet, error) {
	// This is where you would put your implementation of ShowByPetID
	return nil, nil
}

func main() {
	e := gin.Default()
	doc, err := BuildDoc(e)
	if err != nil {
		panic(err)
	}

	if doc.Err() != nil {
		panic(doc.Err())
	}

	bs, err := doc.OpenAPI.Render()
	if err != nil {
		panic(err)
	}

	// Outputs the OpenAPI spec as YAML
	fmt.Println(string(bs))

	// Now serves the API
	err = e.Run(":8080")
	if err != nil {
		panic(err)
	}
}

func BuildDoc(r gin.IRoutes) (*arrestGin.Document, error) {
	baseDoc, err := arrest.NewDocument("Swagger Petstore")
	if err != nil {
		return nil, err
	}

	baseDoc.Version("1.0.0")

	doc := arrestGin.NewDocument(baseDoc, r)

	doc.AddServer("http://petstore.swagger.io/v1")
	doc.PackageMap(
		"pet.v1", "github.com/zostay/arrest-go/gin/examples/petstore/handler",
		"pet.v1", "command-line-arguments",
	)

	listPets := arrest.ParametersFromReflect(reflect.TypeOf(ListPets)).
		P(0, func(p *arrest.Parameter) {
			p.Name("limit").In("query").Required()
		})

	doc.Get("/pets").
		Handler(handleListPets).
		Summary("List all pets").
		OperationID("listPets").
		Tags("pets").
		Parameters(listPets).
		Response("200", func(r *arrest.Response) {
			r.Description("A list of pets").
				Header("x-next", arrest.ModelFrom[string](baseDoc), func(h *arrest.Header) {
					h.Description("A link to the next page of responses")
				}).
				Content("application/json", arrest.ModelFrom[Pets](baseDoc))
		}).
		Response("default", func(r *arrest.Response) {
			r.Description("unexpected error").
				Content("application/json", arrest.ModelFrom[Error](baseDoc))
		})

	doc.Post("/pets").
		Handler(handleCreatePets).
		Summary("Create a pet").
		OperationID("createPets").
		Tags("pets").
		RequestBody("application/json", arrest.ModelFrom[CreatePetRequest](baseDoc)).
		Response("201", func(r *arrest.Response) { r.Description("Null response") }).
		Response("default", func(r *arrest.Response) {
			r.Description("unexpected error").
				Content("application/json", arrest.ModelFrom[Error](baseDoc))
		})

	showByPetId := arrest.ParametersFromReflect(reflect.TypeOf(ShowByPetID)).
		P(0, func(p *arrest.Parameter) {
			p.Name("petId").In("path").Required().
				Description("The ID of the pet to retrieve")
		})

	doc.Get("/pets/{petId}").
		Handler(handleShowByPetID).
		Summary("Info for a specific pet").
		OperationID("showByPetId").
		Tags("pets").
		Parameters(showByPetId).
		Response("200", func(r *arrest.Response) {
			r.Description("Expected response to a valid request").
				Content("application/json", arrest.ModelFrom[Pet](baseDoc))
		}).
		Response("default", func(r *arrest.Response) {
			r.Description("unexpected error").
				Content("application/json", arrest.ModelFrom[Error](baseDoc))
		})

	if doc.Err() != nil {
		return nil, doc.Err()
	}

	return doc, nil
}
