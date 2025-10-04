package gin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	gin2 "github.com/gin-gonic/gin"

	"github.com/zostay/arrest-go"
)

// ErrorResponse represents the standard error response format.
type ErrorResponse struct {
	Status  string            `json:"status"`           // always "error"
	Type    string            `json:"type"`             // error type, e.g. "validation", "internal", etc.
	Message string            `json:"message"`          // general error message
	Fields  map[string]string `json:"fields,omitempty"` // validation messages per field
}

// Error implements the error interface for ErrorResponse.
func (e *ErrorResponse) Error() string {
	return e.Message
}

// validateControllerSignature validates that the controller function has the expected signature:
// func(ctx context.Context, input T) (output U, error)
func (o *Operation) validateControllerSignature(controllerType reflect.Type) error {
	if controllerType.Kind() != reflect.Func {
		return fmt.Errorf("controller must be a function")
	}

	// Check number of parameters and return values
	if controllerType.NumIn() != 2 {
		return fmt.Errorf("controller must have exactly 2 parameters (context.Context, input), got %d", controllerType.NumIn())
	}
	if controllerType.NumOut() != 2 {
		return fmt.Errorf("controller must have exactly 2 return values (output, error), got %d", controllerType.NumOut())
	}

	// Check first parameter is context.Context
	contextType := reflect.TypeOf((*context.Context)(nil)).Elem()
	if !controllerType.In(0).Implements(contextType) {
		return fmt.Errorf("first parameter must implement context.Context, got %s", controllerType.In(0))
	}

	// Check second return value is error
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	if !controllerType.Out(1).Implements(errorType) {
		return fmt.Errorf("second return value must implement error, got %s", controllerType.Out(1))
	}

	return nil
}

// configureOperationSchemas configures the underlying arrest.Operation with request/response schemas
// based on the controller function's input and output types.
func (o *Operation) configureOperationSchemas(inputType, outputType reflect.Type, options *callOptions) error {
	// Configure parameters for fields with In tags (path and query parameters)
	if hasParameterFields(inputType) {
		parameters := arrest.ParametersFromReflect(inputType)
		o.postProcessParameters(parameters, inputType)
		o.Parameters(parameters)
	}

	// Configure request body based on HTTP method and input type
	if o.method != http.MethodGet && o.method != http.MethodDelete {
		// For POST, PUT, etc., use the input type as request body (unless it has only query/path params)
		if hasBodyFields(inputType) {
			// Use ModelFromReflect since we have the reflect.Type
			inputModel := arrest.ModelFromReflect(inputType)
			o.RequestBody("application/json", inputModel)
		}
	}

	// Store output model for later use if needed
	outputModel := arrest.ModelFromReflect(outputType)
	_ = outputModel // TODO: Make this available for manual response configuration

	// Configure error response
	var errorModel *arrest.Model
	if len(options.errorModels) > 0 {
		// Use the first custom error model if provided
		errorModel = options.errorModels[0]
		// TODO: Implement OneOf functionality for combining multiple error models
	} else {
		// Use default error model
		errorModel = arrest.ModelFrom[ErrorResponse]()
	}

	o.Response("default", func(r *arrest.Response) {
		// Use different descriptions based on error model type
		description := "Error"
		if len(options.errorModels) > 0 {
			description = "unexpected error"
		}
		r.Description(description).
			Content("application/json", errorModel)
	})

	return nil
}

// postProcessParameters post-processes generated parameters to handle additional properties like required
// and filters out parameters that don't have explicit in= tags
func (o *Operation) postProcessParameters(parameters *arrest.Parameters, inputType reflect.Type) {
	if inputType.Kind() == reflect.Ptr {
		inputType = inputType.Elem()
	}
	if inputType.Kind() != reflect.Struct {
		return
	}

	// Create a map of field names to their info for quick lookup
	fieldInfo := make(map[string]*arrest.TagInfo)
	for i := 0; i < inputType.NumField(); i++ {
		field := inputType.Field(i)
		info := arrest.NewTagInfo(field.Tag)
		if info.IsIgnored() {
			continue
		}

		fieldName := field.Name
		if info.HasName() {
			fieldName = info.Name()
		}
		fieldInfo[fieldName] = info
	}

	// Filter parameters to only include those with explicit in= tags
	filteredParams := make([]*arrest.Parameter, 0)
	for _, param := range parameters.Parameters {
		if info, exists := fieldInfo[param.Parameter.Name]; exists && info.HasIn() {
			// Update parameter properties
			openAPITag := info.Props()
			if _, hasRequired := openAPITag["required"]; hasRequired {
				param.Required()
			}
			filteredParams = append(filteredParams, param)
		}
	}
	parameters.Parameters = filteredParams
}

// hasParameterFields checks if the input type has fields with In tags (path or query parameters).
func hasParameterFields(inputType reflect.Type) bool {
	if inputType.Kind() == reflect.Ptr {
		inputType = inputType.Elem()
	}
	if inputType.Kind() != reflect.Struct {
		return false // non-struct types don't have parameters
	}

	for i := 0; i < inputType.NumField(); i++ {
		field := inputType.Field(i)
		openAPITag := field.Tag.Get("openapi")

		// Check if field has in=query or in=path tag
		if strings.Contains(openAPITag, "in=query") || strings.Contains(openAPITag, "in=path") {
			return true
		}
	}
	return false
}

// hasBodyFields checks if the input type has fields that should be sent in the request body
// (i.e., fields without openapi:",in=query" or openapi:",in=path" tags).
func hasBodyFields(inputType reflect.Type) bool {
	if inputType.Kind() == reflect.Ptr {
		inputType = inputType.Elem()
	}
	if inputType.Kind() != reflect.Struct {
		return true // non-struct types go in body
	}

	for i := 0; i < inputType.NumField(); i++ {
		field := inputType.Field(i)
		openAPITag := field.Tag.Get("openapi")

		// If no openapi tag or doesn't specify in=query/path, it's a body field
		if !strings.Contains(openAPITag, "in=query") && !strings.Contains(openAPITag, "in=path") {
			return true
		}
	}
	return false
}

// generateHandler creates a gin.HandlerFunc that maps HTTP requests to controller function calls.
func (o *Operation) generateHandler(controller interface{}, inputType, outputType reflect.Type, options *callOptions) gin2.HandlerFunc {
	controllerValue := reflect.ValueOf(controller)

	return func(c *gin2.Context) {
		// Set up panic recovery if requested
		if options.panicProtection {
			defer func() {
				if r := recover(); r != nil {
					err := ErrorResponse{
						Status:  "error",
						Type:    "internal",
						Message: fmt.Sprintf("Internal server error: %v", r),
					}
					c.JSON(http.StatusInternalServerError, err)
				}
			}()
		}

		// Extract input from request
		input, err := o.extractInput(c, inputType)
		if err != nil {
			errResp := ErrorResponse{
				Status:  "error",
				Type:    "validation",
				Message: err.Error(),
			}
			c.JSON(http.StatusBadRequest, errResp)
			return
		}

		// Call the controller function
		results := controllerValue.Call([]reflect.Value{
			reflect.ValueOf(c.Request.Context()),
			input,
		})

		// Handle the results
		output := results[0]
		errValue := results[1]

		// Check for error
		if !errValue.IsNil() {
			err := errValue.Interface().(error)
			errResp := ErrorResponse{
				Status:  "error",
				Type:    "internal",
				Message: err.Error(),
			}
			c.JSON(http.StatusInternalServerError, errResp)
			return
		}

		// Return success response
		c.JSON(http.StatusOK, output.Interface())
	}
}

// extractInput extracts and validates input from the HTTP request based on the input type.
func (o *Operation) extractInput(c *gin2.Context, inputType reflect.Type) (reflect.Value, error) {
	// Create a new instance of the input type
	input := reflect.New(inputType)
	inputElem := input.Elem()

	if inputType.Kind() == reflect.Ptr {
		// If the input type is a pointer, we need to create the pointed-to type
		pointedType := inputType.Elem()
		pointedValue := reflect.New(pointedType)
		input = pointedValue
		inputElem = pointedValue.Elem()
		inputType = pointedType
	}

	if inputType.Kind() != reflect.Struct {
		// For non-struct types, handle as request body
		if o.method != http.MethodGet && o.method != http.MethodDelete {
			if err := c.ShouldBindJSON(input.Interface()); err != nil {
				return reflect.Value{}, fmt.Errorf("failed to parse request body: %w", err)
			}
		}
		return input.Elem(), nil
	}

	// For struct types, extract fields from various sources
	for i := 0; i < inputType.NumField(); i++ {
		field := inputType.Field(i)
		fieldValue := inputElem.Field(i)

		if !fieldValue.CanSet() {
			continue // Skip unexported fields
		}

		// Check openapi and json tags to determine source
		openAPITag := field.Tag.Get("openapi")
		jsonTag := field.Tag.Get("json")

		// Extract field name (use json tag if available, otherwise struct field name)
		fieldName := field.Name
		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" && parts[0] != "-" {
				fieldName = parts[0]
			}
		}

		var value string
		var found bool

		// Determine where to get the value from
		if strings.Contains(openAPITag, "in=path") {
			value = c.Param(fieldName)
			found = value != ""
		} else if strings.Contains(openAPITag, "in=query") || o.method == http.MethodGet || o.method == http.MethodDelete {
			value = c.Query(fieldName)
			found = value != ""
		} else {
			// This field should come from request body
			continue
		}

		// Convert and set the value
		if found {
			if err := o.setFieldValue(fieldValue, value, field.Type); err != nil {
				return reflect.Value{}, fmt.Errorf("failed to convert field %s: %w", fieldName, err)
			}
		}
	}

	// Handle request body fields (those not marked as path or query parameters)
	if o.method != http.MethodGet && o.method != http.MethodDelete && hasBodyFields(inputType) {
		// Create a temporary struct to hold only body fields
		bodyData := make(map[string]interface{})
		if err := c.ShouldBindJSON(&bodyData); err != nil {
			return reflect.Value{}, fmt.Errorf("failed to parse request body: %w", err)
		}

		// Set body fields
		for i := 0; i < inputType.NumField(); i++ {
			field := inputType.Field(i)
			fieldValue := inputElem.Field(i)

			if !fieldValue.CanSet() {
				continue
			}

			openAPITag := field.Tag.Get("openapi")
			jsonTag := field.Tag.Get("json")

			// Skip path and query parameters
			if strings.Contains(openAPITag, "in=path") || strings.Contains(openAPITag, "in=query") {
				continue
			}

			// Extract field name
			fieldName := field.Name
			if jsonTag != "" {
				parts := strings.Split(jsonTag, ",")
				if parts[0] != "" && parts[0] != "-" {
					fieldName = parts[0]
				}
			}

			if bodyValue, exists := bodyData[fieldName]; exists {
				if err := o.setFieldValueFromInterface(fieldValue, bodyValue, field.Type); err != nil {
					return reflect.Value{}, fmt.Errorf("failed to set body field %s: %w", fieldName, err)
				}
			}
		}
	}

	return inputElem, nil
}

// setFieldValue converts a string value to the appropriate type and sets it on the field.
func (o *Operation) setFieldValue(fieldValue reflect.Value, value string, fieldType reflect.Type) error {
	// Handle pointer types
	if fieldType.Kind() == reflect.Ptr {
		if value == "" {
			return nil // Leave nil for empty values
		}
		// Create a new instance of the pointed-to type
		newValue := reflect.New(fieldType.Elem())
		if err := o.setFieldValue(newValue.Elem(), value, fieldType.Elem()); err != nil {
			return err
		}
		fieldValue.Set(newValue)
		return nil
	}

	switch fieldType.Kind() {
	case reflect.String:
		fieldValue.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot convert %q to %s: %w", value, fieldType, err)
		}
		fieldValue.SetInt(intVal)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot convert %q to %s: %w", value, fieldType, err)
		}
		fieldValue.SetUint(uintVal)
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("cannot convert %q to %s: %w", value, fieldType, err)
		}
		fieldValue.SetFloat(floatVal)
	case reflect.Bool:
		boolVal := value == "true" || value == "1"
		fieldValue.SetBool(boolVal)
	default:
		// For time.Time and other custom types, try JSON unmarshaling
		if fieldType == reflect.TypeOf(time.Time{}) {
			t, err := time.Parse(time.RFC3339, value)
			if err != nil {
				return fmt.Errorf("cannot parse time %q: %w", value, err)
			}
			fieldValue.Set(reflect.ValueOf(t))
		} else {
			// Try JSON unmarshaling for complex types
			jsonValue := `"` + value + `"`
			tempValue := reflect.New(fieldType)
			if err := json.Unmarshal([]byte(jsonValue), tempValue.Interface()); err != nil {
				return fmt.Errorf("cannot unmarshal %q to %s: %w", value, fieldType, err)
			}
			fieldValue.Set(tempValue.Elem())
		}
	}

	return nil
}

// setFieldValueFromInterface converts an interface{} value to the appropriate type and sets it on the field.
func (o *Operation) setFieldValueFromInterface(fieldValue reflect.Value, value interface{}, fieldType reflect.Type) error {
	if value == nil {
		return nil
	}

	valueReflect := reflect.ValueOf(value)

	// Handle pointer types
	if fieldType.Kind() == reflect.Ptr {
		newValue := reflect.New(fieldType.Elem())
		if err := o.setFieldValueFromInterface(newValue.Elem(), value, fieldType.Elem()); err != nil {
			return err
		}
		fieldValue.Set(newValue)
		return nil
	}

	// Try direct assignment if types match
	if valueReflect.Type().AssignableTo(fieldType) {
		fieldValue.Set(valueReflect)
		return nil
	}

	// Try conversion if types are convertible
	if valueReflect.Type().ConvertibleTo(fieldType) {
		fieldValue.Set(valueReflect.Convert(fieldType))
		return nil
	}

	// For complex cases, use JSON marshaling/unmarshaling
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cannot marshal value: %w", err)
	}

	tempValue := reflect.New(fieldType)
	if err := json.Unmarshal(jsonBytes, tempValue.Interface()); err != nil {
		return fmt.Errorf("cannot unmarshal to %s: %w", fieldType, err)
	}

	fieldValue.Set(tempValue.Elem())
	return nil
}

// withErr is a helper function to add errors to operations while maintaining chainability.
func withErr[T arrest.ErrHandler](e T, errs ...error) T {
	e.AddError(errs...)
	return e
}
