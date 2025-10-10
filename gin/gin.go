// Package gin provides helpers for integrating arrest-go with the Gin-Gonic web framework.
package gin

import (
	"fmt"
	"net/http"
	"reflect"
	"regexp"

	"github.com/gin-gonic/gin"

	"github.com/zostay/arrest-go"
)

// Document provides a variation on the arrest.Document that helps with route
// registration in a Gin-Gonic router. It wraps an arrest.Document and a Gin IRoutes.
type Document struct {
	*arrest.Document
	r gin.IRoutes
}

// NewDocument creates a new Document for Gin-Gonic route registration.
func NewDocument(doc *arrest.Document, r gin.IRoutes) *Document {
	return &Document{
		Document: doc,
		r:        r,
	}
}

// Get creates a GET operation for the given pattern and returns an Operation for further configuration.
func (d *Document) Get(pattern string) *Operation {
	return &Operation{
		Operation: *d.Document.Get(pattern),
		Document:  d.Document,
		method:    http.MethodGet,
		pattern:   pattern,
		r:         d.r,
	}
}

// Post creates a POST operation for the given pattern and returns an Operation for further configuration.
func (d *Document) Post(pattern string) *Operation {
	return &Operation{
		Operation: *d.Document.Post(pattern),
		Document:  d.Document,
		method:    http.MethodPost,
		pattern:   pattern,
		r:         d.r,
	}
}

// Put creates a PUT operation for the given pattern and returns an Operation for further configuration.
func (d *Document) Put(pattern string) *Operation {
	return &Operation{
		Operation: *d.Document.Put(pattern),
		Document:  d.Document,
		method:    http.MethodPut,
		pattern:   pattern,
		r:         d.r,
	}
}

// Delete creates a DELETE operation for the given pattern and returns an Operation for further configuration.
func (d *Document) Delete(pattern string) *Operation {
	return &Operation{
		Operation: *d.Document.Delete(pattern),
		Document:  d.Document,
		method:    http.MethodDelete,
		pattern:   pattern,
		r:         d.r,
	}
}

// Operation wraps an arrest.Operation and provides Gin-specific route registration methods.
type Operation struct {
	arrest.Operation
	Document *arrest.Document
	method   string
	pattern  string
	r        gin.IRoutes
}

// paramRegex matches OpenAPI-style path parameters (e.g., {id}).
var paramRegex = regexp.MustCompile(`\{([^}]+)\}`)

// patternString translates the OpenAPI spec paths into Gin-Gonic path patterns.
// For example, /foo/{bar} becomes foo/:bar.
func (o *Operation) patternString() string {
	pattern := o.pattern
	if len(pattern) == 0 {
		return pattern
	}

	for pattern[0] == '/' {
		pattern = pattern[1:]
	}

	pattern = paramRegex.ReplaceAllStringFunc(pattern, func(s string) string {
		return ":" + s[1:len(s)-1]
	})

	return pattern
}

// Handler registers a Gin handler for this operation's method and pattern.
func (o *Operation) Handler(handler gin.HandlerFunc) *Operation {
	o.r.Match([]string{o.method}, o.patternString(), handler)
	return o
}

// StaticFile serves a static file for this operation's pattern.
func (o *Operation) StaticFile(file string) *Operation {
	o.r.StaticFile(o.patternString(), file)
	return o
}

// Summary sets the summary for the operation and returns the gin Operation for chaining.
func (o *Operation) Summary(summary string) *Operation {
	o.Operation.Summary(summary)
	return o
}

// Description sets the description for the operation and returns the gin Operation for chaining.
func (o *Operation) Description(description string) *Operation {
	o.Operation.Description(description)
	return o
}

// OperationID sets the operation ID for the operation and returns the gin Operation for chaining.
func (o *Operation) OperationID(operationID string) *Operation {
	o.Operation.OperationID(operationID)
	return o
}

// Tags sets the tags for the operation and returns the gin Operation for chaining.
func (o *Operation) Tags(tags ...string) *Operation {
	o.Operation.Tags(tags...)
	return o
}

// Deprecated marks the operation as deprecated.
func (o *Operation) Deprecated() *Operation {
	o.Operation.Deprecated()
	return o
}

// Call automatically generates a handler for the given controller function.
// The controller must have the signature: func(ctx context.Context, input T) (output U, error)
// where T is the input type and U is the output type.
func (o *Operation) Call(controller interface{}, opts ...CallOption) *Operation {
	// Apply options
	options := &callOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// Validate controller function signature
	controllerType := reflect.TypeOf(controller)
	if err := o.validateControllerSignature(controllerType); err != nil {
		return withErr(o, fmt.Errorf("invalid controller signature: %w", err))
	}

	// Extract input and output types
	inputType := controllerType.In(1)   // Second parameter (first is context.Context)
	outputType := controllerType.Out(0) // First return value (second is error)

	// Configure the operation with inferred schemas
	if err := o.configureOperationSchemas(inputType, outputType, options); err != nil {
		return withErr(o, fmt.Errorf("failed to configure operation schemas: %w", err))
	}

	// Generate and register the handler
	handler := o.generateHandler(controller, inputType, outputType, options)
	o.r.Match([]string{o.method}, o.patternString(), handler)

	return o
}

// CallOption configures behavior for the Call method.
type CallOption func(*callOptions)

// callOptions holds configuration for the Call method.
type callOptions struct {
	errorModels     []*arrest.Model
	panicProtection bool
}

// WithCallErrorModel adds a custom error model to the operation.
// Multiple models will be combined using arrest.OneOf().
func WithCallErrorModel(errModel *arrest.Model) CallOption {
	return func(o *callOptions) {
		o.errorModels = append(o.errorModels, errModel)
	}
}

// WithPanicProtection enables panic protection in the generated handler.
// When enabled, panics are caught and converted to 500 errors instead of crashing.
func WithPanicProtection() CallOption {
	return func(o *callOptions) {
		o.panicProtection = true
	}
}
