package arrest

import (
	"context"
	"errors"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

// Document providees DSL methods for creating OpenAPI documents.
type Document struct {
	// OpenAPI is the underlying OpenAPI document.
	OpenAPI libopenapi.Document

	// DataModel is the v3 DataModel from the document.
	DataModel *libopenapi.DocumentModel[v3.Document]

	// PackageMap maps OpenAPI "package names" to Go package names. This is
	// used in SchemaComponentRef.
	PkgMap map[string]string

	ErrHelper
}

// NewDocumentFromBytes creates a new Document from raw YAML bytes.
func NewDocumentFromBytes(bs []byte) (*Document, error) {
	doc, err := libopenapi.NewDocument(bs)
	if err != nil {
		return nil, err
	}

	return NewDocumentFrom(doc)
}

// NewDocumentFrom creates a new Document from a v3.Document. This allows you
// to add to an existing document using the DSL.
func NewDocumentFrom(doc libopenapi.Document) (*Document, error) {
	dm, errs := doc.BuildV3Model()
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return &Document{
		OpenAPI:   doc,
		DataModel: dm,
	}, nil
}

// NewDocument creates a new Document with the given title.
func NewDocument(title string) (*Document, error) {
	doc := &v3.Document{
		Version: "3.1.0",
		Info: &base.Info{
			Title: title,
		},
	}

	bs, err := doc.Render()
	if err != nil {
		return nil, err
	}

	return NewDocumentFromBytes(bs)
}

func (d *Document) Refresh() error {
	_, _, dm, errs := d.OpenAPI.RenderAndReload()
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	d.DataModel = dm

	return nil
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
	if d.DataModel.Model.Paths == nil {
		d.DataModel.Model.Paths = &v3.Paths{}
	}

	if d.DataModel.Model.Paths.PathItems == nil {
		d.DataModel.Model.Paths.PathItems = orderedmap.New[string, *v3.PathItem]()
	}

	pis := d.DataModel.Model.Paths.PathItems
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
	if d.DataModel.Model.Servers == nil {
		d.DataModel.Model.Servers = []*v3.Server{}
	}

	d.DataModel.Model.Servers = append(d.DataModel.Model.Servers, &v3.Server{URL: url})
	return d
}

// AddSecurityRequirement configures the global security scopes. The key in
// the map is the security scheme name and the value is the list of scopes.
func (d *Document) AddSecurityRequirement(reqs map[string][]string) *Document {
	m := d.DataModel.Model
	if m.Security == nil {
		m.Security = []*base.SecurityRequirement{}
	}

	m.Security = append(m.Security, &base.SecurityRequirement{
		Requirements: orderedmap.ToOrderedMap(reqs),
	})

	return d
}

// SchemaComponent adds a schema component to the document. You can then use
//
//	arrest.SchemaRef(fqn)
//
// to reference this schema in other parts of the document.
func (d *Document) SchemaComponent(fqn string, m *Model) *Document {
	if d.DataModel.Model.Components == nil {
		d.DataModel.Model.Components = &v3.Components{}
	}

	c := d.DataModel.Model.Components
	if c.Schemas == nil {
		c.Schemas = orderedmap.New[string, *base.SchemaProxy]()
	}

	c.Schemas.Set(fqn, m.SchemaProxy)

	return d
}

// SecuritySchemeComponent adds a security scheme component to the document. You
// can then use the fqn to reference this schema in other parts of the document.
func (d *Document) SecuritySchemeComponent(fqn string, m *SecurityScheme) *Document {
	if d.DataModel.Model.Components == nil {
		d.DataModel.Model.Components = &v3.Components{}
	}

	c := d.DataModel.Model.Components
	if c.SecuritySchemes == nil {
		c.SecuritySchemes = orderedmap.New[string, *v3.SecurityScheme]()
	}

	c.SecuritySchemes.Set(fqn, m.SecurityScheme)

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
	if d.DataModel.Model.Components == nil {
		return nil
	}

	if d.DataModel.Model.Components.Schemas == nil {
		return nil
	}

	scs := make([]*SchemaComponent, 0, d.DataModel.Model.Components.Schemas.Len())
	for pair := range orderedmap.Iterate(ctx, d.DataModel.Model.Components.Schemas) {
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
	if d.DataModel.Model.Paths == nil {
		return nil
	}

	if d.DataModel.Model.Paths.PathItems == nil {
		return nil
	}

	os := make([]*Operation, 0, d.DataModel.Model.Paths.PathItems.Len())
	for pair := range orderedmap.Iterate(ctx, d.DataModel.Model.Paths.PathItems) {
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
