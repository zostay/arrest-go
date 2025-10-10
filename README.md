# Arrest Go!

The OpenAPI 3.1 spec generator and REST interface handler for Go. This
is built on the very excellent [libopenapi](https://github.com/pb33f/libopenapi)
library from [pb33f](https://pb33f.io/). Five out of five
stars. Highly recommend. It can handle just about any kind of weird, fancy, or
crazy OpenAPI doc you want to read or write, but it's also like programming your
VCR.

**Arrest Go** provides both a powerful DSL for building OpenAPI specifications
and automatic handler generation for popular the [Gin](https://gin-gonic.com)
web frameworks. Whether you want fine-grained control over your API
documentation or prefer convention-over-configuration with automatic shims,
Arrest Go has you covered.

# Why?

Honestly, the state of OpenAPI generation in Go is not super great, so while
there's a vacuum in support for OpenAPI 3.1, I thought I'd give it a try. On the
other hand, OpenAPI 3.1 library support is fantastic because pb33f and quobix
and friends are pretty much the best. This is built on top of his library and I
highly recommend all of their applications and libraries, so three cheers to
Dave and his misfit engineers.

The general goal here is for something that:

* **Works!** - Reliable, production-ready OpenAPI generation
* **Outputs modern OpenAPI 3.1** - Support for the latest OpenAPI
  specification
* **No external gateway** - Doesn't require some external gateway...
  seriously, why is this a thing?
* **Automatic handler generation** - Creates server shims automatically with
  framework integration (only Gin for now)
* **Type-safe** - Reuse your own Go types for schemas and method definitions
* **Flexible** - Both high-level automation and low-level control when needed

It is still early days and a little experimental, but it is used in production.

# Quick Start

## 🚀 Automatic Handler Generation (Recommended)

The fastest way to get started is using Gin with the Call method, which
automatically generates configuration and HTTP handlers to map requests to
controller logic and back to responses.

Here's the Pet Store example:

```go
package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zostay/arrest-go"
	arrestgin "github.com/zostay/arrest-go/gin"
)

// Define your data types
type Pet struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Tag  string `json:"tag"`
}

type Pets []Pet

// Define request/response types with parameter tags
type CreatePetRequest struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Tag  string `json:"tag"`
	Type string `json:"type" openapi:",in=query"` // Query parameter
}

type GetPetRequest struct {
	PetId string `json:"petId" openapi:",in=path"` // Path parameter
}

type PetListRequest struct {
	Type  string `json:"type" openapi:",in=query"`           // Optional query param
	Limit int32  `json:"limit" openapi:",in=query,required"` // Required query param
}

// Write your controller functions
func CreatePet(ctx context.Context, req CreatePetRequest) (*Pet, error) {
	// Your business logic here
	pet := &Pet{ID: 1, Name: req.Name, Tag: req.Tag}
	return pet, nil
}

func GetPet(ctx context.Context, req GetPetRequest) (*Pet, error) {
	// Your business logic here
	return &Pet{ID: 1, Name: "Fluffy", Tag: "cat"}, nil
}

func ListPets(ctx context.Context, req PetListRequest) (Pets, error) {
	// Your business logic here
	return Pets{{ID: 1, Name: "Fluffy", Tag: "cat"}}, nil
}

func main() {
	// Create arrest document
	arrestDoc, err := arrest.NewDocument("Pet Store API")
	if err != nil {
		log.Fatal(err)
	}

	arrestDoc.Version("1.0.0").
		Description("A simple pet store API").
		AddServer("http://localhost:8080")

	// Create gin router and arrestgin document
	router := gin.Default()
	doc := arrestgin.NewDocument(arrestDoc, router)

	// Define operations using Call method - handlers are generated automatically!
	doc.Get("/pets").
		OperationID("listPets").
		Tags("pets").
		Summary("List all pets").
		Call(ListPets) // 🎉 Automatic handler generation

	doc.Post("/pets").
		OperationID("createPets").
		Tags("pets").
		Summary("Create a pet").
		Call(CreatePet).
		Response("201", func(r *arrest.Response) {
			r.Description("Pet created successfully")
		})

	doc.Get("/pets/{petId}").
		OperationID("getPet").
		Tags("pets").
		Summary("Get a pet by ID").
		Call(GetPet)

	// Add OpenAPI spec endpoint
	router.GET("/openapi.yaml", func(c *gin.Context) {
		spec, err := arrestDoc.OpenAPI.Render()
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.Header("Content-Type", "application/yaml")
		c.String(200, string(spec))
	})

	log.Println("🚀 Server starting on :8080")
	log.Println("📖 OpenAPI spec: http://localhost:8080/openapi.yaml")
	router.Run(":8080")
}
```

### What the Call Method Does

The `Call()` method automatically:

- ✅ **Validates function signatures**
- ✅ **Extracts parameters** from struct tags (`openapi:",in=path"`,
  `openapi:",in=query"`)
- ✅ **Generates request body schemas** from non-parameter fields
- ✅ **Creates response schemas** from return types
- ✅ **Registers HTTP handlers** with your web framework
- ✅ **Handles parameter binding** (path, query, body)
- ✅ **Provides error handling** with standardized responses

With this setup, you can also use all the other features of Arrest Go, to control
the details of your OpenAPI spec as needed. This library will work to help ensure
that your code matches the spec in the process (with the caveat that some of the
validators are still WIP).

### Struct Tag Reference

| Tag                            | Description              | Example                                                    |
|--------------------------------|--------------------------|------------------------------------------------------------|
| `openapi:",in=query"`          | Query parameter          | `Limit int32 \`json:"limit" openapi:",in=query"\``         |
| `openapi:",in=path"`           | Path parameter           | `ID string \`json:"id" openapi:",in=path"\``               |
| `openapi:",in=query,required"` | Required query parameter | `Name string \`json:"name" openapi:",in=query,required"\`` |
| `openapi:"-"`                  | Exclude field completely | `Internal string \`openapi:"-"\``                          |
| No tag                         | Request body field       | `Name string \`json:"name"\``                              |

## 🛠️ Manual Handler Example

If, instead, you want to write your own handlers or integrate with a framework
other than Gin. you can still use this to define your OpenAPI doc:

```go
package main

import (
	"fmt"
	"reflect"

	"github.com/zostay/arrest-go"
)

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

func main() {
	doc, err := arrest.NewDocument("Swagger Petstore")
	if err != nil {
		panic(err)
	}

	doc.Version("1.0.0").
		Description("A sample API that uses petstore as an example").
		AddServer("http://petstore.swagger.io/v1")

	// Manual parameter definition with full control
	listPetsParams := arrest.NParameters(1).
		P(0, func(p *arrest.Parameter) {
			p.Name("limit").InQuery().Required().
				Description("maximum number of results to return").
				Model(arrest.ModelFrom[int32](doc))
		})

	doc.Get("/pets").
		Summary("List all pets").
		OperationID("listPets").
		Tags("pets").
		Parameters(listPetsParams).
		Response("200", func(r *arrest.Response) {
			r.Description("A list of pets").
				Header("x-next", arrest.ModelFrom[string](doc), func(h *arrest.Header) {
					h.Description("A link to the next page of responses")
				}).
				Content("application/json", arrest.ModelFrom[Pets](doc))
		}).
		Response("default", func(r *arrest.Response) {
			r.Description("unexpected error").
				Content("application/json", arrest.ModelFrom[Error](doc))
		})

	doc.Post("/pets").
		Summary("Create a pet").
		OperationID("createPets").
		Tags("pets").
		RequestBody("application/json", arrest.ModelFrom[Pet](doc)).
		Response("201", func(r *arrest.Response) {
			r.Description("Null response")
		}).
		Response("default", func(r *arrest.Response) {
			r.Description("unexpected error").
				Content("application/json", arrest.ModelFrom[Error](doc))
		})

	petIdParam := arrest.NParameters(1).
		P(0, func(p *arrest.Parameter) {
			p.Name("petId").InPath().Required().
				Description("The ID of the pet to retrieve").
				Model(arrest.ModelFrom[string](doc))
		})

	doc.Get("/pets/{petId}").
		Summary("Info for a specific pet").
		OperationID("showByPetId").
		Tags("pets").
		Parameters(petIdParam).
		Response("200", func(r *arrest.Response) {
			r.Description("Expected response to a valid request").
				Content("application/json", arrest.ModelFrom[Pet](doc))
		}).
		Response("default", func(r *arrest.Response) {
			r.Description("unexpected error").
				Content("application/json", arrest.ModelFrom[Error](doc))
		})

	if doc.Err() != nil {
		panic(doc.Err())
	}

	rend, err := doc.OpenAPI.Render()
	if err != nil {
		panic(err)
	}

	fmt.Print(string(rend))
}
```

## 🔄 Loading and Modifying Existing Documents

You can also load an existing OpenAPI document and modify it programmatically. This is useful for enhancing existing specs or migrating from other tools:

```go
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/pb33f/libopenapi"
	"github.com/zostay/arrest-go"
)

func main() {
	// Load an existing OpenAPI document
	yamlSpec := `
openapi: 3.1.0
info:
  title: Existing API
  version: 1.0.0
paths:
  /users:
    get:
      summary: Get users
      responses:
        '200':
          description: Success
`

	// Create a ReadCloser from the YAML string
	reader := io.NopCloser(strings.NewReader(yamlSpec))

	// Parse the document using libopenapi
	oas, err := libopenapi.NewDocument(reader)
	if err != nil {
		log.Fatal("Failed to parse document:", err)
	}

	// Create an Arrest Go document from the existing OpenAPI document
	doc := arrest.NewDocument(oas)
	if err := doc.Err(); err != nil {
		log.Fatal("Failed to create arrest document:", err)
	}

	// Modify existing operations
	ctx := context.Background()
	operations := doc.Operations(ctx)
	for _, op := range operations {
		// Add a new response to existing operations
		op.Response("400", func(r *arrest.Response) {
			r.Description("Bad Request").
				Content("application/json", arrest.ModelFrom[map[string]string](doc))
		})

		// Add tags to existing operations
		op.Tags("users", "api")

		// Set operation ID if not present
		if op.Operation.OperationId == "" {
			op.OperationID("getUsers")
		}
	}

	// Add entirely new operations
	type CreateUserRequest struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	type User struct {
		ID    int64  `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	doc.Post("/users").
		Summary("Create a new user").
		OperationID("createUser").
		Tags("users", "api").
		RequestBody("application/json", arrest.ModelFrom[CreateUserRequest](doc)).
		Response("201", func(r *arrest.Response) {
			r.Description("User created successfully").
				Content("application/json", arrest.ModelFrom[User](doc))
		}).
		Response("400", func(r *arrest.Response) {
			r.Description("Invalid input").
				Content("application/json", arrest.ModelFrom[map[string]string](doc))
		})

	// Add new paths completely
	doc.Get("/users/{id}").
		Summary("Get user by ID").
		OperationID("getUserById").
		Tags("users", "api").
		Parameters(
			arrest.NParameters(1).
				P(0, func(p *arrest.Parameter) {
					p.Name("id").InPath().Required().
						Description("User ID").
						Model(arrest.ModelFrom[int64](doc))
				}),
		).
		Response("200", func(r *arrest.Response) {
			r.Description("User found").
				Content("application/json", arrest.ModelFrom[User](doc))
		}).
		Response("404", func(r *arrest.Response) {
			r.Description("User not found").
				Content("application/json", arrest.ModelFrom[map[string]string](doc))
		})

	// Update document metadata
	doc.Version("2.0.0").
		Description("Enhanced API with additional endpoints").
		AddServer("https://api.example.com/v2")

	// Check for errors
	if doc.Err() != nil {
		log.Fatal("Document errors:", doc.Err())
	}

	// Render the modified document
	rendered, err := doc.OpenAPI.Render()
	if err != nil {
		log.Fatal("Failed to render document:", err)
	}

	fmt.Print(string(rendered))
}
```

### Loading from File

You can also load documents from files:

```go
func loadFromFile(filename string) (*arrest.Document, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Parse with libopenapi
	oas, err := libopenapi.NewDocument(file)
	if err != nil {
		return nil, err
	}

	// Create arrest document
	doc := arrest.NewDocument(oas)
	return doc, doc.Err()
}

// Usage
doc, err := loadFromFile("existing-api.yaml")
if err != nil {
	log.Fatal(err)
}

// Now modify the document as needed...
```

This approach is particularly useful for:
- **Migration scenarios**: Converting from other OpenAPI tools
- **API versioning**: Creating new versions of existing APIs
- **Documentation enhancement**: Adding missing details to existing specs
- **Specification merging**: Combining multiple API documents

# 🎯 Features

## Core Capabilities

### 📝 **Full OpenAPI 3.1 Support**

- Complete OpenAPI 3.1 specification compliance
- OAS 3.1 JSON Schema support

### 🔄 **Automatic Code Generation**

- **Handler Generation**: Automatically create HTTP handlers from controller
  functions (for Gin)
- **Parameter Extraction**: Automatic path and query parameter binding
- **Request/Response Binding**: Type-safe JSON marshaling and unmarshaling
- **Schema Inference**: Generate OpenAPI schemas from Go types
- **Customization**: Customize serialization handling via JSON marshaler and unmarshaler interfaces

### 🏗️ **Flexible Architecture**

- **Manual DSL**: Complete control over every aspect of your API specification
- **Automatic Mode**: Use your existing Go types and logic functions to generate specs
- **Hybrid Approach**: Mix automatic generation with manual overrides
- **Edit Existing Specs**: Load and modify existing OpenAPI documents
- **Attach to Existing Specs**: If you already have an OpenAPI doc, you can attach handlers to it while you migrate
- **Framework Integration**: First-class support for popular Go web frameworks

### 🛡️ **Type Safety**

- Compile-time validation of controller function signatures (partially working, but also WIP)
- Strong typing for all OpenAPI components
- Automatic schema validation

### 🔧 **Developer Experience**

- Method chaining DSL for readable API definitions
- Inferrence from Go types to reduce boilerplate and duplication
- Comprehensive error reporting with context
- Hot-reload support for development
- Automatic documentation generation from your Godoc struct and field comments

## Framework Integration Features

### Gin Integration

```go
import arrestgin "github.com/zostay/arrest-go/gin"

// Automatic handler registration
doc := arrestgin.NewDocument(arrestDoc, ginRouter)
doc.Get("/api/users").Call(GetUsers) // Handler registered automatically!
```

**Features:**

- ✅ Automatic route registration
- ✅ Parameter binding (path, query, body)
- ✅ Error handling
- ✅ Request validation

### Call Method Options

```go
doc.Post("/pets").
    Call(CreatePet,
    WithCallErrorModel(customErrorModel), // Custom error responses
)
```

## Advanced Features

### 🔄 **Recursive Type Handling**

Automatic detection and handling of self-referencing and mutually recursive
types:

```go
type User struct {
    ID       int64   `json:"id"`
    Friends  []User  `json:"friends"` // Self-referencing - handled automatically!
    Profile  Profile `json:"profile"`
}

type Profile struct {
    UserID int64 `json:"user_id"`
    User   *User `json:"user"` // Mutual recursion - handled automatically!
}
```

### 📦 **Package Mapping**

Clean schema names with package mapping:

```go
doc.PackageMap("api.v1", "github.com/company/api/v1")
// Results in clean schema names like "User" instead of "github.com.company.api.v1.User"
```

### 🏷️ **Rich Tagging System**

Comprehensive struct tag support. Will work with "json" and "openapi" tags:

```go
type SearchRequest struct {
	// Search query
    Query    string   `json:"q" openapi:",in=query,required"`
	//Filter by tags
    Tags     []string `json:"tags" openapi:",in=query"`
    UserID   int64    `json:"user_id" openapi:",in=path,required"`
    Internal string   `json:"-" openapi:"-"` // Excluded completely
    Metadata struct {
        Source string `json:"source"`
    } `json:"metadata"` // Nested in request body
}
```

The aim is to speed up building your specs while giving you full flexibility.
This tool operates at a high level, but you can always drop down to "hard mode"
and manipulate the underlying `OpenAPI` struct directly when needed. The 
libopenapi library gives you direct control of the YAML node objects used to
work with the specs directly, which is very powerful (but comes with a learning
curve and foot guns, of course).

# The DSL

That's a Domain Specific Language for those who may not know. This library is
based around building a sort of DSL inside of Go. It makes heavy use of method
chaining to help abbreviate the code necessary to build a spec. This takes a
little getting used.

The chaining is setup to be used in paragraphs or stanzas around the primary
components. We consider the following pieces to be primary components:

* Document
* Operation

Therefore, the encouraged style is to build up the Document and then build up
each Operation. Each primary component returns the component object and then all
the methods of the primary only return that object.

```go
doc := arrest.NewDocument("My API").AddServer("http://example.com")

doc.Get("/path").Summary("Get a thing") // and on

doc.Post("/path").Summary("Post a thing") // and on
```

Within each primary there are secondary components. To avoid breaking up your
primary paragraphs, these do not return the different types for configuration
but make use of callbacks or inner constructors.

```go
doc.Get("/path").Summary("Get a thing").
    OperationID("getThing").
    Parameters(
        arrest.ParametersFromReflect(reflect.TypeOf(GetThing)).
            P(0, func (p *arrest.Parameter) {
                p.Name("id").In("query").Required()
            })
    ).
    Response("200", func (r *arrest.Response) {
        r.Description("The thing").
            Content("application/json", arrest.ModelFrom[Thing](doc))
    })
```

Finally, when you're finished you need to check to see if errors occurred. The
system should not panic, but it records errors as it goes. You can check for
them at the end, or any time in the middle by calling the `Err()` method on the
document. You can check for errors in operations or other parts as well. Errors
all flow upward to the parent component, but you can check them at any level:

```go
if doc.Err() != nil {
    // handle the error here, of course
}
```

# 📦 Installation

## Core Library

```shell
go get github.com/zostay/arrest-go
```

## Gin Framework Integrations

```shell
# For Gin framework support
go get github.com/zostay/arrest-go/gin

# Add to your go.mod
require (
    github.com/zostay/arrest-go v0.x.x
    github.com/zostay/arrest-go/gin v0.x.x  // For Gin integration
)
```

## Quick Setup

```bash
# Clone and explore examples
git clone https://github.com/zostay/arrest-go.git
cd arrest-go

# Run the basic example
go run examples/petstore.go

# Run the Gin Call method example
cd gin && go run examples/petstore/call/petstore.go
# Visit http://localhost:8080/openapi.yaml to see the generated spec!

# Run tests
go test ./...           # Core library tests
cd gin && go test ./... # Gin integration tests
```

# 🤝 Contributing

We welcome contributions! Here's how to get started:

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **Make your changes** and add tests
4. **Run tests**: `go test ./...` and `cd gin && go test ./...`
5. **Submit a pull request**

## Development Setup

```bash
git clone https://github.com/zostay/arrest-go.git
cd arrest-go

# Install dependencies
go mod download
cd gin && go mod download

# Run all tests
cd .. && go test ./...
cd gin && go test ./...

# Run examples to verify functionality
go run examples/petstore.go
cd gin && go run examples/petstore/call/petstore.go
```

# 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file
for details.

# 🙏 Special Thanks

Thanks to [pb33f](https://pb33f.io/) for the
excellent [libopenapi](https://github.com/pb33f/libopenapi) library. This
library is built on top of that one.

Also thanks to [Adrian Hesketh](https://github.com/a-h)
whose [github.com/a-h/rest](https://github.com/a-h/rest) library provided some
inspiration for this library.

---

**Star ⭐ this repo if you find it useful!** It helps others discover the project
and motivates continued development.
