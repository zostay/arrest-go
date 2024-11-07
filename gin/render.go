package gin

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"text/template"
	"unicode"

	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/zostay/arrest-go"
)

const (
	DefaultPkgName      = "rest"
	DefaultServiceName  = "Service"
	DefaultEndpointName = "Endpoint"
)

//go:embed templates/service.go.tmpl
var serviceTemplate string

// Gin is a code generator that will output a set of RESTful routes to handlers
// which then call a service interface to perform the actual work.
type Gin struct {
	*arrest.Document

	PkgName     string
	ServiceName string

	opCounter int
}

type param struct {
	GoName string
	WireName string
	Type string
	In string
}

type handler struct {
	Name   string
	Input  []param
	Output []param

	Method string
	Path   string

	Body string
}

type renderVars struct {
	PkgName     string
	ServiceName string
	Handlers    []handler
}

func (g *Gin) Generate(w io.Writer) error {
	t, err := template.New("service").ParseFiles(serviceTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	v := renderVars{
		PkgName:     g.pkgName(),
		ServiceName: g.serviceName(),
		Handlers:    g.handlers(),
	}

	err = t.Execute(w, g)
}

func (g *Gin) pkgName() string {
	if g.PkgName != "" {
		return g.PkgName
	}
	return DefaultPkgName
}

func cleanIdentifier(in string) string {
	first := true
	out := make([]byte, 0, len(in))
	for _, r := range in {
		if first && (!unicode.IsLetter(r) && r != '_') {
			continue
		}

		first = false

		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			out = append(out, byte(r))
		}
	}

	return string(out)
}

func (g *Gin) serviceName() string {
	if g.ServiceName != "" {
		return g.ServiceName
	}

	svcName := cleanIdentifier(g.Document.DataModel.Model.Info.Title)
	if svcName != "" {
		return svcName
	}
	return DefaultServiceName
}

func (g *Gin) handlers() []handler {
	os := g.Document.Operations(context.TODO())
	handlers := make([]handler, 0, len(os))
	for _, o := range os {
		handlers = append(handlers, handler{
			Name:  g.operationName(o.Operation),
			Input: g.operationInput(o.Operation),
		})
	}

	return handlers
}

func (g *Gin) operationName(op *v3.Operation) string {
	name := cleanIdentifier(op.OperationId)
	if name != "" {
		return name
	}

	name = fmt.Sprint(DefaultEndpointName, g.opCounter)
	g.opCounter++
	return name
}

func (g *Gin) operationInput(op *v3.Operation) []param {
	params := make([]param, 0, len(op.Parameters)+1)
	for _, p := range op.Parameters {
		goType := ""
		goToType := ""
		goFromType := ""

		types := p.Schema.Schema().Type
		if len(types) > 0 {
			panic("multiple types not implemented")
		}

		for _, t := range types {
			switch t {
			case "integer":
				goType = "int"
				goToType = "{{.Tmp}}, {{.Err}} := strconv.ParseInt({{.In}}, 10, 64)"
				goFromType = "strconv.FormatInt(%s, 10)"
			case "number":
				goType = "float64"
				goToType = "float64(strconv.ParseFloat(%s, 64))"

			case "boolean":
				goType = "bool"
			default:
				panic("parameter type " + t + " not implemented")
			}
		}

		params = append(params, param{
			Name: p.Name,
			Type: goType,
			ConvertFromString:
		})
	}

	return params
}
