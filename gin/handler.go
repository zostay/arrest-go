package gin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"time"

	gin2 "github.com/gin-gonic/gin"

	"github.com/zostay/arrest-go"
)

// HTTPStatusCoder is an interface for objects that can provide their own HTTP status code.
type HTTPStatusCoder interface {
	HTTPStatusCode() int
}

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
			if options.requestComponent {
				// For component references, use the underlying type (strip pointer)
				componentType := inputType
				if componentType.Kind() == reflect.Ptr {
					componentType = componentType.Elem()
				}
				inputModel := arrest.ModelFromReflect(componentType, o.Document, arrest.AsComponent())
				inputRef := arrest.SchemaRef(inputModel.MappedName(o.Document.PkgMap))
				o.RequestBody("application/json", inputRef)
			} else {
				inputModel := arrest.ModelFromReflect(inputType, o.Document)
				o.RequestBody("application/json", inputModel)
			}
		}
	}

	// Configure success response with output model
	// Only add if no responses have been configured yet
	hasAnyResponse := o.Operation.Operation.Responses != nil &&
		o.Operation.Operation.Responses.Codes != nil &&
		o.Operation.Operation.Responses.Codes.Len() > 0

	if !hasAnyResponse {
		o.Response("200", func(r *arrest.Response) {
			// Get description from output model's godoc
			description := "Success"
			actualType := outputType
			if outputType.Kind() == reflect.Ptr {
				actualType = outputType.Elem()
			}
			if actualType.PkgPath() != "" && actualType.Name() != "" {
				if godocComment := arrest.GoDocForType(actualType); godocComment != "" {
					description = godocComment
				}
			}

			if options.responseComponent {
				// For component references, use the underlying type (strip pointer)
				componentType := outputType
				if componentType.Kind() == reflect.Ptr {
					componentType = componentType.Elem()
				}
				outputModel := arrest.ModelFromReflect(componentType, o.Document, arrest.AsComponent())
				outputRef := arrest.SchemaRef(outputModel.MappedName(o.Document.PkgMap))
				r.Description(description).
					Content("application/json", outputRef)
			} else {
				outputModel := arrest.ModelFromReflect(outputType, o.Document)
				r.Description(description).
					Content("application/json", outputModel)
			}
		})
	}

	// Configure error response
	var errorModel *arrest.Model
	if len(options.replaceErrorModels) > 0 {
		// Replace default error models completely with custom error models
		if len(options.replaceErrorModels) == 1 {
			// Use single replace error model
			errorModel = options.replaceErrorModels[0]
		} else {
			// Combine multiple replace error models using OneOf
			errorModel = arrest.OneOfTheseModels(o.Document, options.replaceErrorModels...)
		}
	} else if len(options.errorModels) > 0 {
		// Always include the default ErrorResponse along with custom error models
		defaultErrorModel := arrest.ModelFrom[ErrorResponse](o.Document)
		allErrorModels := make([]*arrest.Model, 0, len(options.errorModels)+1)
		allErrorModels = append(allErrorModels, defaultErrorModel)
		allErrorModels = append(allErrorModels, options.errorModels...)

		// Combine default and custom error models using OneOf
		errorModel = arrest.OneOfTheseModels(o.Document, allErrorModels...)
	} else {
		// Use default error model
		errorModel = arrest.ModelFrom[ErrorResponse](o.Document)
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
		info := arrest.NewTagInfo(field.Tag)

		// Check if field has in=query or in=path tag
		if info.In() == "query" || info.In() == "path" {
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
		info := arrest.NewTagInfo(field.Tag)

		// If no openapi tag or doesn't specify in=query/path, it's a body field
		if info.In() != "query" && info.In() != "path" {
			return true
		}
	}
	return false
}

// defaultErrorHandler is the default error handler that checks if the error is an *ErrorResponse
// and returns it as-is if so. Otherwise, it constructs an internal server error.
// If the original error implements HTTPStatusCoder, it preserves that interface.
func defaultErrorHandler(ctx *gin2.Context, err error) interface{} {
	// Check if the error is already an *ErrorResponse
	if errResp, ok := err.(*ErrorResponse); ok {
		return errResp
	}

	// If the error implements HTTPStatusCoder, return it as-is to preserve the status code
	if _, ok := err.(HTTPStatusCoder); ok {
		return err
	}

	// Construct a new ErrorResponse for other error types
	return &ErrorResponse{
		Status:  "error",
		Type:    "internal",
		Message: err.Error(),
	}
}

// handleError processes an error using the appropriate error handler and returns the response with status code.
// It handles the custom error handler selection and HTTPStatusCoder interface checking.
func (o *Operation) handleError(c *gin2.Context, err error, options *callOptions, defaultStatusCode int) (interface{}, int) {
	// Use custom error handler if provided, otherwise use default
	var errorHandler ErrorHandlerFunc
	if options.errorHandler != nil {
		errorHandler = options.errorHandler
	} else {
		errorHandler = defaultErrorHandler
	}

	errResp := errorHandler(c, err)

	// Check if the error response implements HTTPStatusCoder
	statusCode := defaultStatusCode
	if statusCoder, ok := errResp.(HTTPStatusCoder); ok {
		statusCode = statusCoder.HTTPStatusCode()
	}

	return errResp, statusCode
}

// generateHandler creates a gin.HandlerFunc that maps HTTP requests to controller function calls.
func (o *Operation) generateHandler(controller interface{}, inputType, outputType reflect.Type, options *callOptions) gin2.HandlerFunc {
	controllerValue := reflect.ValueOf(controller)

	return func(c *gin2.Context) {
		// Set up panic recovery if requested
		if options.panicProtection {
			defer func() {
				if r := recover(); r != nil {
					panicErr := &ErrorResponse{
						Status:  "error",
						Type:    "internal",
						Message: fmt.Sprintf("Internal server error: %v", r),
					}
					errResp, statusCode := o.handleError(c, panicErr, options, http.StatusInternalServerError)
					c.JSON(statusCode, errResp)
				}
			}()
		}

		// Extract input from request
		input, err := o.extractInput(c, inputType)
		if err != nil {
			validationErr := &ErrorResponse{
				Status:  "error",
				Type:    "validation",
				Message: err.Error(),
			}
			errResp, statusCode := o.handleError(c, validationErr, options, http.StatusBadRequest)
			c.JSON(statusCode, errResp)
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
			errResp, statusCode := o.handleError(c, err, options, http.StatusInternalServerError)
			c.JSON(statusCode, errResp)
			return
		}

		// Return success response
		outputInterface := output.Interface()

		// Check if the response implements HTTPStatusCoder
		statusCode := http.StatusOK
		if statusCoder, ok := outputInterface.(HTTPStatusCoder); ok {
			statusCode = statusCoder.HTTPStatusCode()
		}

		c.JSON(statusCode, outputInterface)
	}
}

// extractInput extracts and validates input from the HTTP request based on the input type.
func (o *Operation) extractInput(c *gin2.Context, inputType reflect.Type) (reflect.Value, error) {
	isPointerType := inputType.Kind() == reflect.Ptr

	// Create a new instance of the input type
	var input reflect.Value
	var inputElem reflect.Value

	if isPointerType {
		// If the input type is a pointer, we need to create the pointed-to type
		pointedType := inputType.Elem()
		pointedValue := reflect.New(pointedType)
		input = pointedValue
		inputElem = pointedValue.Elem()
		inputType = pointedType
	} else {
		// For non-pointer types, create a pointer to the type and work with the element
		input = reflect.New(inputType)
		inputElem = input.Elem()
	}

	if inputType.Kind() != reflect.Struct {
		// For non-struct types, handle as request body
		if o.method != http.MethodGet && o.method != http.MethodDelete {
			if err := c.ShouldBindJSON(input.Interface()); err != nil {
				return reflect.Value{}, fmt.Errorf("failed to parse request body: %w", err)
			}
		}
		if isPointerType {
			return input, nil
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

		info := arrest.NewTagInfo(field.Tag)

		// Extract field name (use json tag if available, otherwise struct field name)
		fieldName := field.Name
		if tagName := info.Name(); tagName != "" {
			fieldName = tagName
		}

		var value string
		var found bool

		// Determine where to get the value from
		if info.In() == "path" {
			value = c.Param(fieldName)
			found = value != ""
		} else if info.In() == "query" || o.method == http.MethodGet || o.method == http.MethodDelete {
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

			info := arrest.NewTagInfo(field.Tag)

			// Skip path and query parameters
			if info.In() == "path" || info.In() == "query" {
				continue
			}

			// Extract field name
			fieldName := field.Name
			if tagName := info.Name(); tagName != "" {
				fieldName = tagName
			}

			if bodyValue, exists := bodyData[fieldName]; exists {
				if err := o.setFieldValueFromInterface(fieldValue, bodyValue, field.Type); err != nil {
					return reflect.Value{}, fmt.Errorf("failed to set body field %s: %w", fieldName, err)
				}
			}
		}
	}

	if isPointerType {
		return input, nil
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
