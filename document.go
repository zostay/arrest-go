package arrest

import (
	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

type Document struct {
	OpenAPI *v3.Document

	ErrHelper
}

func NewDocumentFrom(v3doc *v3.Document) *Document {
	return &Document{OpenAPI: v3doc}
}

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

func (d *Document) AddServer(url string) *Document {
	if d.OpenAPI.Servers == nil {
		d.OpenAPI.Servers = []*v3.Server{}
	}

	d.OpenAPI.Servers = append(d.OpenAPI.Servers, &v3.Server{URL: url})
	return d
}