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

### Recent Changes
Recent work has focused on:
- Implementing the Call method for automatic handler generation from controller functions
- Improving recursive type handling and ensuring package mapping is correctly applied to all schema references
- Adding automatic parameter extraction and request body inference for gin operations
- Enhancing error handling with standardized error response formats

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

### Testing Gin Features
```bash
cd gin && go test ./...                    # Test all gin functionality
cd gin && go test ./examples/petstore/call # Test Call method example
```