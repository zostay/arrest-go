package gin

import (
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/zostay/arrest-go"
)

// Document provides a variation on the arrest.Document that helps with route
// registration in a Gin-Gonic router.
type Document struct {
	*arrest.Document
	r gin.IRoutes
}

func NewDocument(doc *arrest.Document, r gin.IRoutes) *Document {
	return &Document{
		Document: doc,
		r:        r,
	}
}

func (d *Document) Get(pattern string) *Operation {
	return &Operation{
		Operation: *d.Document.Get(pattern),
		method:    http.MethodGet,
		pattern:   pattern,
		r:         d.r,
	}
}

func (d *Document) Post(pattern string) *Operation {
	return &Operation{
		Operation: *d.Document.Post(pattern),
		method:    http.MethodPost,
		pattern:   pattern,
		r:         d.r,
	}
}

func (d *Document) Put(pattern string) *Operation {
	return &Operation{
		Operation: *d.Document.Put(pattern),
		method:    http.MethodPut,
		pattern:   pattern,
		r:         d.r,
	}
}

func (d *Document) Delete(pattern string) *Operation {
	return &Operation{
		Operation: *d.Document.Delete(pattern),
		method:    http.MethodDelete,
		pattern:   pattern,
		r:         d.r,
	}
}

type Operation struct {
	arrest.Operation
	method  string
	pattern string
	r       gin.IRoutes
}

var paramRegex = regexp.MustCompile(`\{([^}]+)\}`)

// patternString translates the OpenAPI spec paths into Gin-Gonic path patterns.
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

func (o *Operation) Handler(handler gin.HandlerFunc) *Operation {
	o.r.Match([]string{o.method}, o.patternString(), handler)
	return o
}

func (o *Operation) StaticFile(file string) *Operation {
	o.r.StaticFile(o.patternString(), file)
	return o
}
