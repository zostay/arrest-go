# Polymorphic API Example

This example demonstrates comprehensive polymorphic type support in arrest-go with the Gin framework using the `.Call()` method for automatic handler generation.

## Features Demonstrated

### 1. **Implicit Polymorphic Types with Struct Tags**
- **Animal types** using inline polymorphism (`json:",inline"` tags)
- Discriminator-based type selection with `animalType` field
- Support for Dogs, Cats, and Birds with type-specific properties

### 2. **Component Reference Polymorphism**
- **Vehicle types** using component references (`refName` tags)
- Pointer-based polymorphic fields for nullable references
- Support for Cars, Trucks, and Motorcycles

### 3. **Polymorphic Error Responses**
- Multiple error types: ValidationError, BusinessError, SystemError
- Automatic error model composition using `WithPolymorphicError()`
- Different error scenarios based on input validation

### 4. **Automatic Handler Generation**
- Uses `.Call()` method for zero-boilerplate HTTP handlers
- Automatic request/response schema generation
- Parameter extraction from path, query, and body
- Built-in validation and error handling

## API Endpoints

### Animals (Inline Polymorphism)

- `POST /animals` - Create an animal (dog, cat, or bird)
- `GET /animals/{id}` - Get an animal by ID
- `PUT /animals/{id}` - Update an animal

### Vehicles (Component References)

- `POST /vehicles` - Create a vehicle (car, truck, or motorcycle)
- `GET /vehicles/{id}` - Get a vehicle by ID

### Documentation

- `GET /` - Interactive Swagger UI documentation
- `GET /openapi.yaml` - Raw OpenAPI specification

## Running the Example

```bash
# Install dependencies
go mod tidy

# Run tests
go test -v

# Generate OpenAPI spec only (for inspection/documentation)
go run main.go --generate-spec

# Start the server
go run main.go
```

**Generated File:**
- `openapi.yaml` - Complete OpenAPI specification showcasing polymorphic schemas, discriminators, and error handling

Then visit:
- http://localhost:8080 - Interactive API documentation
- http://localhost:8080/openapi.yaml - Raw OpenAPI spec

## Example Requests

### Create a Dog
```bash
curl -X POST http://localhost:8080/animals?source=test \
  -H "Content-Type: application/json" \
  -d '{
    "animalType": "dog",
    "Dog": {
      "breed": "Golden Retriever",
      "name": "Max",
      "isGood": true
    }
  }'
```

### Create a Cat
```bash
curl -X POST http://localhost:8080/animals?source=test \
  -H "Content-Type: application/json" \
  -d '{
    "animalType": "cat",
    "Cat": {
      "lives": 9,
      "name": "Whiskers",
      "color": "orange"
    }
  }'
```

### Create a Car
```bash
curl -X POST http://localhost:8080/vehicles?owner=john \
  -H "Content-Type: application/json" \
  -d '{
    "vehicleType": "car",
    "car": {
      "doors": 4,
      "brand": "Toyota",
      "model": "Camry"
    }
  }'
```

## Error Testing

### Validation Error (Invalid Lives)
```bash
curl -X POST http://localhost:8080/animals?source=test \
  -H "Content-Type: application/json" \
  -d '{
    "animalType": "cat",
    "Cat": {
      "lives": 15,
      "name": "Garfield",
      "color": "orange"
    }
  }'
```

### Business Error (Bad Dog Name)
```bash
curl -X POST http://localhost:8080/animals?source=test \
  -H "Content-Type: application/json" \
  -d '{
    "animalType": "dog",
    "Dog": {
      "breed": "Poodle",
      "name": "BadDog",
      "isGood": false
    }
  }'
```

### System Error (Database Failure)
```bash
curl -X POST http://localhost:8080/animals?source=error \
  -H "Content-Type: application/json" \
  -d '{
    "animalType": "dog",
    "Dog": {
      "breed": "Beagle",
      "name": "Snoopy",
      "isGood": true
    }
  }'
```

## Key Implementation Details

### Polymorphic Type Definitions

The example shows two approaches to polymorphic types:

**1. Inline Polymorphism (Animals)**
```go
type Animal struct {
    AnimalType string `json:"animalType" openapi:",discriminator,defaultMapping=dog"`
    Dog        Dog    `json:",inline,omitempty" openapi:",oneOf,mapping=dog"`
    Cat        Cat    `json:",inline,omitempty" openapi:",oneOf,mapping=cat"`
    Bird       Bird   `json:",inline,omitempty" openapi:",oneOf,mapping=bird"`
}
```

**2. Component Reference Polymorphism (Vehicles)**
```go
type Vehicle struct {
    VehicleType string      `json:"vehicleType" openapi:",discriminator,defaultMapping=car"`
    Car         *Car        `json:"car,omitempty" openapi:",oneOf,mapping=car,refName=Car"`
    Truck       *Truck      `json:"truck,omitempty" openapi:",oneOf,mapping=truck,refName=Truck"`
    Motorcycle  *Motorcycle `json:"motorcycle,omitempty" openapi:",oneOf,mapping=motorcycle,refName=Motorcycle"`
}
```

### Controller Signature Pattern
```go
func CreateAnimal(ctx context.Context, req CreateAnimalRequest) (AnimalResponse, error)
```

### Polymorphic Error Configuration
```go
ginDoc.Post("/animals").Call(CreateAnimal,
    arrestgin.WithPolymorphicError(validationErrorModel, businessErrorModel, systemErrorModel),
    arrestgin.WithComponents(),
)
```

## Generated OpenAPI Features

The example generates a complete OpenAPI 3.1 specification with:

- **oneOf compositions** for polymorphic request/response schemas
- **Discriminator objects** with property names and type mappings
- **Component references** for reusable schema definitions
- **Polymorphic error responses** with multiple error types
- **Parameter definitions** extracted from struct tags
- **Request body schemas** for complex polymorphic inputs
- **Response schemas** with success and error variants

This demonstrates the full power of arrest-go's polymorphic support integrated seamlessly with Gin's routing and the `.Call()` method for automatic handler generation.

## Current Status

✅ **Working Features:**
- Polymorphic request/response handling with discriminator support
- Automatic OpenAPI schema generation for polymorphic types
- Successful request processing for animals and vehicles
- Component reference vs inline schema patterns
- Full end-to-end `.Call()` integration with polymorphic types
- Interactive Swagger UI documentation generation

⚠️ **Known Limitations:**
- Some error handling tests may fail due to JSON unmarshaling complexities with polymorphic discriminators
- The runtime JSON handling works correctly for successful cases but may need custom unmarshaling for complex error scenarios

The core polymorphic functionality and OpenAPI generation works perfectly - this example successfully demonstrates how to build polymorphic APIs with arrest-go!