// Package gin provides helpers for integrating arrest-go with the Gin-Gonic web framework.
package gin

import (
	"net/http"
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
		method:    http.MethodGet,
		pattern:   pattern,
		r:         d.r,
	}
}

// Post creates a POST operation for the given pattern and returns an Operation for further configuration.
func (d *Document) Post(pattern string) *Operation {
	return &Operation{
		Operation: *d.Document.Post(pattern),
		method:    http.MethodPost,
		pattern:   pattern,
		r:         d.r,
	}
}

// Put creates a PUT operation for the given pattern and returns an Operation for further configuration.
func (d *Document) Put(pattern string) *Operation {
	return &Operation{
		Operation: *d.Document.Put(pattern),
		method:    http.MethodPut,
		pattern:   pattern,
		r:         d.r,
	}
}

// Delete creates a DELETE operation for the given pattern and returns an Operation for further configuration.
func (d *Document) Delete(pattern string) *Operation {
	return &Operation{
		Operation: *d.Document.Delete(pattern),
		method:    http.MethodDelete,
		pattern:   pattern,
		r:         d.r,
	}
}

// Operation wraps an arrest.Operation and provides Gin-specific route registration methods.
type Operation struct {
	arrest.Operation
	method  string
	pattern string
	r       gin.IRoutes
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
