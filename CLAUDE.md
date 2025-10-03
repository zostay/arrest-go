# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

AR! Rest! is a Go library that provides a Domain Specific Language (DSL) for generating OpenAPI 3.0/3.1 specifications. Built on top of pb33f/libopenapi, it emphasizes method chaining, reflection-based schema generation, and type safety.

## Common Commands

### Testing
```bash
go test ./...                    # Run all tests in project
go test -v ./...                 # Verbose test output  
go test -race ./...              # Test with race detection
go test -run TestModelFrom       # Run specific test patterns
go test ./gin/...                # Test gin integration only
```

### Building
```bash
go build ./...                   # Build all packages
go mod tidy                      # Clean up dependencies
go run examples/petstore.go      # Run main example
go run gin/examples/petstore/rest/petstore.go  # Run gin example
```

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

### Schema Reference Management
When working with schema components:
- Child schemas are automatically extracted via `ExtractChildRefs()`
- Package mapping is applied during component registration in `SchemaComponent()`
- The `remapSchemaRefs()` function recursively updates all `$ref` values in schemas

### Recent Changes
Recent work has focused on improving recursive type handling and ensuring package mapping is correctly applied to all schema references, including those generated during recursive type processing.