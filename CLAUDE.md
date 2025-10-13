# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

AR! Rest! is a Go library that provides a Domain Specific Language (DSL) for generating OpenAPI 3.0/3.1 specifications. Built on top of pb33f/libopenapi, it emphasizes method chaining, reflection-based schema generation, and type safety.

## Common Commands

### Quick Development (using Makefile)
```bash
make help                        # Show all available targets
make test                        # Run tests in both root and gin modules
make quick-check                 # Format, vet, and test (fast dev cycle)
make ci                          # Run full CI pipeline locally
make dev-setup                   # Setup development environment
```

### Testing
```bash
make test                        # Run all tests in both modules
make test-verbose                # Verbose test output
make test-race                   # Test with race detection
make test-coverage               # Test with coverage reports
make bench                       # Run benchmarks
go test -run TestModelFrom       # Run specific test patterns (manual)
```

### Building
```bash
make build                       # Build all packages in both modules
make clean                       # Clean build artifacts and caches
make examples                    # Run all example programs
```

### Code Quality
```bash
make fmt                         # Format code in both modules
make vet                         # Run go vet in both modules
make lint                        # Run golangci-lint in both modules
make check                       # Run fmt + vet + lint
```

### Dependency Management
```bash
make mod-tidy                    # Run go mod tidy in both modules
make mod-verify                  # Verify modules
make mod-download                # Download dependencies
scripts/retidy-pr <branch-name>  # Run go mod tidy on a PR branch using git worktrees
scripts/retidy-prs               # Run retidy-pr on all PRs with failed tests
scripts/retidy-pr --help         # Show retidy-pr usage information
scripts/retidy-prs --help        # Show retidy-prs usage information
```

The `retidy-pr` script is useful for handling individual Dependabot PRs that need `go mod tidy` synchronization across multiple modules. The `retidy-prs` script automates this by processing all open PRs with failed tests.

## Architecture Overview

### Core Components
- **Document**: Root DSL component managing OpenAPI document structure and component registry
- **Operation**: HTTP method operations (GET, POST, etc.) with parameter, request body, and response configuration
- **Model**: Go type reflection engine that converts Go structs to OpenAPI schemas with recursive type handling
- **Component**: Schema component wrapper managing references and package mapping

### DSL Design Philosophy
The library uses a two-tier chaining pattern:
- **Primary Components** (Document, Operation): Return themselves to enable method chaining
- **Secondary Components** (Response, Parameter): Use callback functions to avoid breaking chain flow

Example:
```go
doc.Get("/pets").Summary("List pets").
    Response("200", func(r *Response) {
        r.Description("Success").Content("application/json", model)
    })
```

### Package Structure
- **Main package** (`arrest`): Pure OpenAPI spec generation
- **Gin subpackage** (`gin/`): Gin framework integration that wraps main package and adds route registration
- **Examples** (`examples/`, `gin/examples/`): Working examples demonstrating usage patterns

## Key Technical Details

### Recursive Type Handling
The library handles self-referencing and mutually recursive Go types through:
- `refMapper` tracks types currently being processed to detect cycles
- Pre-registration of named struct types in component registry  
- Automatic `$ref` generation when recursion is detected
- Post-processing to apply package mapping to all schema references

### Error Handling
All components embed `ErrHelper` for hierarchical error propagation. Always check `doc.Err()` after building specifications.

### Package Mapping
The `PackageMap` system allows mapping Go package paths to OpenAPI schema names:
```go
doc.PackageMap("api.v1", "github.com/company/api/v1")
```

### Schema Generation
- Uses reflection to convert Go types to OpenAPI schemas
- Supports `json` and `openapi` struct tags for customization
- Extracts Go documentation for OpenAPI descriptions when `SkipDocumentation = false`
- Automatically handles complex types (slices, maps, pointers, embedded structs)

## Testing Patterns

Tests follow these conventions:
- Golden file testing with `assert.YAMLEq()` for comparing generated OpenAPI output
- Use `t.Parallel()` for independent tests
- Integration tests build complete documents and verify rendering
- All test functions should check `doc.Err()` before proceeding

### Test Structure Example
```go
func TestModelFrom_Example(t *testing.T) {
    t.Parallel()
    
    doc, err := arrest.NewDocument("test")
    require.NoError(t, err)
    
    // Configure document
    doc.Get("/path").Response("200", func(r *Response) {
        r.Content("application/json", model)
    })
    
    assert.NoError(t, doc.Err())
    oas, err := doc.OpenAPI.Render()
    require.NoError(t, err)
    assert.YAMLEq(t, expected, string(oas))
}
```

## Important Implementation Notes

### Method Chaining Requirements
- Primary components must always return themselves from configuration methods
- Secondary components use callback patterns to maintain primary component chains
- Error collection happens via embedded `ErrHelper` - never panic during DSL operations

### Gin Integration Pattern
The gin subpackage wraps the main package and adds route registration. Operations return enhanced versions that can register HTTP handlers while maintaining the same DSL interface.

#### Call Method (Arrest Shims)
The gin package provides a `Call()` method that automatically generates HTTP handlers from controller functions:

```go
// Controller function signature: func(ctx context.Context, input T) (output U, error)
func CreatePet(ctx context.Context, req CreatePetRequest) (*Pet, error) {
    // Implementation
}

// Automatic handler generation with Call method
doc.Post("/pets").
    OperationID("createPets").
    Tags("pets").
    Summary("Create a pet").
    Call(CreatePet)  // Automatically generates handler, parameters, request body, and responses
```

**Key Features:**
- **Automatic Parameter Generation**: Extracts path and query parameters from struct fields with `openapi:",in=path"` or `openapi:",in=query"` tags
- **Request Body Inference**: Fields without parameter tags become request body properties
- **Response Schema Generation**: Automatically creates OpenAPI schemas from return types
- **Signature Validation**: Validates controller function signatures at compile time
- **Error Handling**: Provides standardized error responses with panic protection option

**Struct Tag Support:**
- `openapi:",in=query"` - Query parameter
- `openapi:",in=path"` - Path parameter
- `openapi:",in=query,required"` - Required query parameter
- `openapi:"-"` - Exclude field completely
- Fields without tags become request body properties

**Options:**
- `WithCallErrorModel(model)` - Custom error response model
- `WithPanicProtection()` - Automatic panic recovery

### Schema Reference Management
When working with schema components:
- Child schemas are automatically extracted via `ExtractChildRefs()`
- Package mapping is applied during component registration in `SchemaComponent()`
- The `remapSchemaRefs()` function recursively updates all `$ref` values in schemas

### Component Registration (Updated API)
**BREAKING CHANGE**: Component registration has been completely redesigned for explicit control:

**OLD (Removed)**:
```go
// This API no longer exists
errRef := doc.SchemaComponentRef(arrest.ModelFrom[Error]()).Ref()
```

**NEW (Current)**:
```go
// Explicit component registration
errModel := arrest.ModelFrom[Error](doc, arrest.AsComponent()).
    Description("An error response.")
errRef := arrest.SchemaRef(errModel.MappedName(doc.PkgMap))

// Or with custom component name
model := arrest.ModelFrom[User](doc, arrest.WithComponentName("UserModel"))
```

**Key Changes**:
- `ModelFrom[T]()` → `ModelFrom[T](doc, opts...)` - Document context now required
- Components only registered when `AsComponent()` option is used
- Child references only registered as components when parent is a component
- Fixed bug where all models were automatically registered as components

### Polymorphic Type Support
arrest-go provides comprehensive support for OpenAPI 3+ polymorphic schemas:

#### Explicit Polymorphic Functions
```go
// Constructor functions for polymorphic compositions
oneOfModel := arrest.OneOfTheseModels(doc, dogModel, catModel, birdModel).
    Discriminator("animalType", "dog", "dog", "cat", "bird")

anyOfModel := arrest.AnyOfTheseModels(doc, mammalModel, birdModel)
allOfModel := arrest.AllOfTheseModels(doc, baseModel, extendedModel)
```

#### Implicit Polymorphic Types via Struct Tags
```go
type Animal struct {
    AnimalType string `json:"animalType" openapi:",discriminator,defaultMapping=dog"`
    Dog        Dog    `json:",inline,omitempty" openapi:",oneOf,mapping=dog"`
    Cat        Cat    `json:",inline,omitempty" openapi:",oneOf,mapping=cat"`
    Bird       Bird   `json:",inline,omitempty" openapi:",oneOf,mapping=bird"`
}

// With component references
type Vehicle struct {
    VehicleType string      `json:"vehicleType" openapi:",discriminator,defaultMapping=car"`
    Car         *Car        `json:"car,omitempty" openapi:",oneOf,mapping=car,refName=Car"`
    Truck       *Truck      `json:"truck,omitempty" openapi:",oneOf,mapping=truck,refName=Truck"`
    Motorcycle  *Motorcycle `json:"motorcycle,omitempty" openapi:",oneOf,mapping=motorcycle,refName=Motorcycle"`
}
```

**Supported Struct Tags:**
- `openapi:",discriminator,defaultMapping=<type>"` - Discriminator property with default
- `openapi:",oneOf,mapping=<type>"` - OneOf variant mapping
- `openapi:",anyOf,mapping=<type>"` - AnyOf variant mapping
- `openapi:",allOf,mapping=<type>"` - AllOf variant mapping
- `openapi:",oneOf,mapping=<type>,refName=<name>"` - Component reference variant

#### Gin Polymorphic Integration
```go
// Polymorphic error responses
ginDoc.Post("/animals").Call(CreateAnimal,
    arrestgin.WithPolymorphicError(validationErrorModel, businessErrorModel, systemErrorModel))

// Automatic polymorphic request/response handling
func CreateAnimal(ctx context.Context, req CreateAnimalRequest) (AnimalResponse, error) {
    // Controller automatically handles polymorphic input based on discriminator
    switch req.AnimalType {
    case "dog":
        return processDog(req.Dog)
    case "cat":
        return processCat(req.Cat)
    }
}
```

### Recent Changes
Recent work has focused on:
- **MAJOR: Polymorphic Type Support**: Complete implementation of OpenAPI 3+ polymorphic schemas (oneOf, anyOf, allOf) with discriminator support
- **Implicit Polymorphic Types**: Struct tag-based polymorphic definition for declarative schema generation
- **Gin Polymorphic Integration**: Full polymorphic support in gin Call method with `WithPolymorphicError()` option
- **Breaking Change**: `ModelFrom` now requires document context: `ModelFrom[T]()` → `ModelFrom[T](doc, opts...)`
- **New Component Registration**: Added `AsComponent()` and `WithComponentName()` options for explicit component registration
- **Fixed Component Registration Bug**: Components are now only registered when explicitly requested with `AsComponent()`
- **Removed `SchemaComponentRef()`**: Replaced with `ModelFrom[T](doc, AsComponent())` + `SchemaRef()`

## Working with the Gin Subpackage

### Directory Navigation
When working on gin-specific functionality, use the following pattern:
```bash
cd gin && command  # Navigate to gin directory first
```
The gin subpackage has its own go.mod and should be treated as a separate module.

### Gin Examples
- `gin/examples/petstore/rest/` - Manual handler implementation
- `gin/examples/petstore/call/` - Call method implementation demonstrating automatic handler generation
- `gin/examples/polymorphic/` - **NEW**: Comprehensive polymorphic API example with discriminator support

### Testing Gin Features
```bash
cd gin && go test ./...                         # Test all gin functionality
cd gin && go test ./examples/petstore/call      # Test Call method example
cd gin && go test ./examples/polymorphic        # Test polymorphic functionality
```

### Running Examples
```bash
cd gin/examples/polymorphic && go run main.go  # Start polymorphic API server
# Visit http://localhost:8080 for interactive Swagger UI
```