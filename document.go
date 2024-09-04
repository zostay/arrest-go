package arrest

import (
	"context"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

// Document providees DSL methods for creating OpenAPI documents.
type Document struct {
	// OpenAPI is the underlying OpenAPI document.
	OpenAPI *v3.Document

	// PackageMap maps OpenAPI "package names" to Go package names. This is
	// used in SchemaComponentRef.
	PkgMap map[string]string

	ErrHelper
}

// NewDocumentFrom creates a new Document from a v3.Document. This allows you
// to add to an existing document using the DSL.
func NewDocumentFrom(v3doc *v3.Document) *Document {
	return &Document{OpenAPI: v3doc}
}

// NewDocument creates a new Document with the given title.
func NewDocument(title string) *Document {
	return &Document{
		OpenAPI: &v3.Document{
			Version: "3.1.0",
			Info: &base.Info{
				Title: title,
			},
		},
	}
}

func (d *Document) PackageMap(pairs ...string) *Document {
	if d.PkgMap == nil {
		d.PkgMap = make(map[string]string)
	}

	for i := 0; i < len(pairs); i += 2 {
		d.PkgMap[pairs[i]] = pairs[i+1]
	}

	return d
}

func (d *Document) pathItem(pattern string) *v3.PathItem {
	if d.OpenAPI.Paths == nil {
		d.OpenAPI.Paths = &v3.Paths{}
	}

	if d.OpenAPI.Paths.PathItems == nil {
		d.OpenAPI.Paths.PathItems = orderedmap.New[string, *v3.PathItem]()
	}

	pis := d.OpenAPI.Paths.PathItems
	if _, hasPi := pis.Get(pattern); !hasPi {
		pis.Set(pattern, &v3.PathItem{})
	}

	return pis.GetOrZero(pattern)
}

// Get creates a new GET operation at the given pattern. The Operation is
// returned to be manipulated further.
func (d *Document) Get(pattern string) *Operation {
	pi := d.pathItem(pattern)

	if pi.Get == nil {
		pi.Get = &v3.Operation{}
	}

	v3o := pi.Get

	o := &Operation{Operation: v3o}
	d.AddHandler(o)
	return o
}

// Post creates a new POST operation at the given pattern. The Operation is
// returned to be manipulated further.
func (d *Document) Post(pattern string) *Operation {
	pi := d.pathItem(pattern)

	if pi.Post == nil {
		pi.Post = &v3.Operation{}
	}

	v3o := pi.Post

	o := &Operation{Operation: v3o}
	d.AddHandler(o)
	return o
}

// AddServer adds a new server URL to the document.
func (d *Document) AddServer(url string) *Document {
	if d.OpenAPI.Servers == nil {
		d.OpenAPI.Servers = []*v3.Server{}
	}

	d.OpenAPI.Servers = append(d.OpenAPI.Servers, &v3.Server{URL: url})
	return d
}

// SchemaComponent adds a schema component to the document. You can then use
//
//	arrest.SchemaRef(fqn)
//
// to reference this schema in other parts of the document.
func (d *Document) SchemaComponent(fqn string, m *Model) *Document {
	if d.OpenAPI.Components == nil {
		d.OpenAPI.Components = &v3.Components{}
	}

	c := d.OpenAPI.Components
	if c.Schemas == nil {
		c.Schemas = orderedmap.New[string, *base.SchemaProxy]()
	}

	c.Schemas.Set(fqn, m.SchemaProxy)

	return d
}

func (d *Document) SchemaComponentRef(m *Model) *SchemaComponent {
	fqn := m.MappedName(d.PkgMap)

	d.SchemaComponent(fqn, m)

	return &SchemaComponent{
		schema: m,
		ref:    SchemaRef(fqn),
	}
}

// SchemaComponents lists all the schema components in the document.
func (d *Document) SchemaComponents(ctx context.Context) []*SchemaComponent {
	if d.OpenAPI.Components == nil {
		return nil
	}

	if d.OpenAPI.Components.Schemas == nil {
		return nil
	}

	scs := make([]*SchemaComponent, 0, d.OpenAPI.Components.Schemas.Len())
	for pair := range orderedmap.Iterate(ctx, d.OpenAPI.Components.Schemas) {
		name, sp := pair.Key(), pair.Value()

		scs = append(scs, &SchemaComponent{
			schema: &Model{
				Name:        name,
				SchemaProxy: sp,
			},
			ref: SchemaRef(name),
		})
	}

	return scs
}

// Operations lists all the operations in the document.
func (d *Document) Operations(ctx context.Context) []*Operation {
	if d.OpenAPI.Paths == nil {
		return nil
	}

	if d.OpenAPI.Paths.PathItems == nil {
		return nil
	}

	os := make([]*Operation, 0, d.OpenAPI.Paths.PathItems.Len())
	for pair := range orderedmap.Iterate(ctx, d.OpenAPI.Paths.PathItems) {
		pi := pair.Value()

		if pi.Get != nil {
			os = append(os, &Operation{Operation: pi.Get})
		}
		if pi.Post != nil {
			os = append(os, &Operation{Operation: pi.Post})
		}
		if pi.Delete != nil {
			os = append(os, &Operation{Operation: pi.Delete})
		}
		if pi.Put != nil {
			os = append(os, &Operation{Operation: pi.Put})
		}
		if pi.Patch != nil {
			os = append(os, &Operation{Operation: pi.Patch})
		}
		if pi.Options != nil {
			os = append(os, &Operation{Operation: pi.Options})
		}
		if pi.Head != nil {
			os = append(os, &Operation{Operation: pi.Head})
		}
		if pi.Trace != nil {
			os = append(os, &Operation{Operation: pi.Trace})
		}
	}

	return os
}
