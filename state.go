package arrest

import (
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/index"
)

// The DSL types keep their libopenapi objects behind untyped fields, and
// every function that touches one is marked //go:noinline. This is not
// style: the Go compiler re-emits every generic method reachable from any
// type a package names — through struct fields, method signatures, and the
// bodies it inlines — into the importing package (golang/go#70511), and
// libopenapi's model reaches thousands of them. A consumer that names
// *arrest.Document would otherwise pay two seconds of compile time for
// every package that does so. See issue #101 and internal/compilecost, whose
// tests enforce the rule.
//
// The helpers here are the only way from a DSL value to its libopenapi
// object. Inside the package, use them; outside it, the OpenAPI* functions
// in raw.go wrap them.

// documentState is what a Document holds.
type documentState struct {
	model *v3.Document
	index *index.SpecIndex
}

// modelState is what a Model holds.
type modelState struct {
	schema        *base.SchemaProxy
	makeRefs      map[string]*base.SchemaProxy
	componentRefs map[string]*base.SchemaProxy // refs that should be registered as components
}

//go:noinline
func docOf(d *Document) *documentState {
	return d.state.(*documentState)
}

//go:noinline
func modelOf(d *Document) *v3.Document {
	return d.state.(*documentState).model
}

//go:noinline
func stateOf(m *Model) *modelState {
	return m.state.(*modelState)
}

//go:noinline
func schemaOf(m *Model) *base.SchemaProxy {
	return m.state.(*modelState).schema
}

//go:noinline
func opOf(o *Operation) *v3.Operation {
	return o.state.(*v3.Operation)
}

//go:noinline
func responseOf(r *Response) *v3.Response {
	return r.state.(*v3.Response)
}

//go:noinline
func paramOf(p *Parameter) *v3.Parameter {
	return p.state.(*v3.Parameter)
}

//go:noinline
func headerOf(h *Header) *v3.Header {
	return h.state.(*v3.Header)
}

//go:noinline
func schemeOf(s *SecurityScheme) *v3.SecurityScheme {
	return s.state.(*v3.SecurityScheme)
}

//go:noinline
func flowsOf(f *RegardingFlow) []*v3.OAuthFlow {
	return f.state.([]*v3.OAuthFlow)
}

//go:noinline
func newModel(name string, sp *base.SchemaProxy, makeRefs, componentRefs map[string]*base.SchemaProxy) *Model {
	return &Model{Name: name, state: &modelState{schema: sp, makeRefs: makeRefs, componentRefs: componentRefs}}
}

//go:noinline
func newOperation(op *v3.Operation) *Operation {
	return &Operation{state: op}
}

//go:noinline
func newParameter(p *v3.Parameter) *Parameter {
	return &Parameter{state: p}
}

//go:noinline
func newSecurityScheme(s *v3.SecurityScheme) *SecurityScheme {
	return &SecurityScheme{state: s}
}
