package arrest

import (
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/index"
)

// The functions in this file reach past the DSL to the libopenapi objects
// underneath, for anything the DSL does not cover.
//
// A package that calls one of them names a libopenapi type, and so pays
// what the DSL types are built not to cost: the compiler emits every generic
// method reachable from that type into the calling package, several thousand
// functions and a second or two of compile time per package (see state.go).
// Keep such code in a package of its own, so that the rest of a program does
// not pay for it.

// OpenAPIDocument returns the libopenapi document the DSL builds.
//
//go:noinline
func OpenAPIDocument(d *Document) *v3.Document {
	return modelOf(d)
}

// DocumentIndex returns the index built when the document was last loaded
// or refreshed.
//
//go:noinline
func DocumentIndex(d *Document) *index.SpecIndex {
	return docOf(d).index
}

// OpenAPISchema returns the schema proxy behind a model.
//
//go:noinline
func OpenAPISchema(m *Model) *base.SchemaProxy {
	return schemaOf(m)
}

// ModelFromOpenAPISchema wraps a libopenapi schema proxy as a model, for use
// anywhere the DSL takes one. Name is the fully qualified Go type name the
// model stands for, or "" for an anonymous schema.
//
//go:noinline
func ModelFromOpenAPISchema(name string, sp *base.SchemaProxy) *Model {
	return newModel(name, sp, map[string]*base.SchemaProxy{}, map[string]*base.SchemaProxy{})
}

// OpenAPIChildRefs returns the schemas a model references by refName or
// elemRefName, keyed by fully qualified Go type name.
//
//go:noinline
func OpenAPIChildRefs(m *Model) map[string]*base.SchemaProxy {
	return stateOf(m).makeRefs
}

// OpenAPIComponentRefs returns the subset of a model's child refs that are
// registered as components when the model is.
//
//go:noinline
func OpenAPIComponentRefs(m *Model) map[string]*base.SchemaProxy {
	return stateOf(m).componentRefs
}

// OpenAPIOperation returns the libopenapi operation behind an Operation.
//
//go:noinline
func OpenAPIOperation(o *Operation) *v3.Operation {
	return opOf(o)
}

// OpenAPIResponse returns the libopenapi response behind a Response.
//
//go:noinline
func OpenAPIResponse(r *Response) *v3.Response {
	return responseOf(r)
}

// OpenAPIParameter returns the libopenapi parameter behind a Parameter.
//
//go:noinline
func OpenAPIParameter(p *Parameter) *v3.Parameter {
	return paramOf(p)
}

// OpenAPIHeader returns the libopenapi header behind a Header.
//
//go:noinline
func OpenAPIHeader(h *Header) *v3.Header {
	return headerOf(h)
}

// OpenAPISecurityScheme returns the libopenapi security scheme behind a
// SecurityScheme.
//
//go:noinline
func OpenAPISecurityScheme(s *SecurityScheme) *v3.SecurityScheme {
	return schemeOf(s)
}

// OpenAPIFlows returns the OAuth flows a RegardingFlow is configuring.
//
//go:noinline
func OpenAPIFlows(f *RegardingFlow) []*v3.OAuthFlow {
	return flowsOf(f)
}
