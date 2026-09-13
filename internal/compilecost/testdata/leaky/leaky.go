// Package leaky breaks the opacity rule in every way Opaque must catch.
package leaky

import (
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// ViaField reaches libopenapi through a field.
type ViaField struct{ M *v3.Document }

// ViaMethod reaches libopenapi through an unexported method's signature.
type ViaMethod struct{ x int }

func (v *ViaMethod) model() *v3.Document { return nil }

// hidden is unexported and tainted; ViaHidden reaches it through a pointer.
type hidden struct{ sp *base.SchemaProxy }

// ViaHidden reaches libopenapi through an unexported struct.
type ViaHidden struct{ h *hidden }

// Clean is fine: an any field, and clean methods.
type Clean struct{ state any }

// Good touches libopenapi and says so.
//
//go:noinline
func (c *Clean) Good() { c.state.(*v3.Document).Info.Title = "ok" }

// BadBody touches libopenapi in its body without saying so.
func (c *Clean) BadBody() { c.state.(*v3.Document).Info.Title = "no" }

// BadSig mentions libopenapi in its signature without saying so.
func BadSig(d *v3.Document) {}

// Untouched is fine.
func Untouched(c *Clean) *Clean { return c }
