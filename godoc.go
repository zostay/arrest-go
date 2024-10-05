package arrest

import (
	"fmt"
	"go/ast"
	"go/doc"
	"reflect"

	"golang.org/x/tools/go/packages"
)

type fieldDoc struct {
	Name    string
	Comment string
	Tag     reflect.StructTag
}

func goDocForFields(spec ast.Spec) map[string]fieldDoc {
	fieldComm := map[string]fieldDoc{}
	if typeSpec, isTypeSpec := spec.(*ast.TypeSpec); isTypeSpec {
		if structType, isStructType := typeSpec.Type.(*ast.StructType); isStructType {
			for _, field := range structType.Fields.List {
				if len(field.Names) == 0 {
					continue
				}

				comment := ""
				if field.Doc != nil {
					comment = field.Doc.Text()
				}

				tag := ""
				if field.Tag != nil {
					tag = field.Tag.Value
				}

				fieldName := field.Names[0].Name
				fieldComm[fieldName] = fieldDoc{
					Name:    fieldName,
					Comment: comment,
					Tag:     reflect.StructTag(tag),
				}
			}
		}
	}

	return fieldComm
}

func GoDocForStruct(t reflect.Type) (string, map[string]string, error) {
	// NOTE: I implemented this in a hurry without really understanding what the
	// hell I'm doing. To quote one of my son's favorite sayings, "Men learn
	// mostly through trial and error, but mostly error." That's exactly what
	// this is. It works for my purpose, but I definitely hacked this together
	// out of spit and baling wire to get it to work.

	if t.Kind() != reflect.Struct {
		return "", nil, fmt.Errorf("expected a struct type, got %s", t.Kind())
	}

	if t.PkgPath() == "" {
		return "", nil, nil
	}

	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedFiles,
	}, t.PkgPath())
	if err != nil {
		return "", nil, err
	}

	if len(pkgs) == 0 {
		return "", nil, nil
	}

	pkg := pkgs[0]
	if pkg.Fset == nil || pkg.Syntax == nil {
		return "", nil, nil
	}

	docPkg, err := doc.NewFromFiles(pkg.Fset, pkg.Syntax, t.PkgPath())
	if err != nil {
		return "", nil, err
	}

	for _, docType := range docPkg.Types {
		if docType.Name == t.Name() {
			comment := docType.Doc

			var fieldMap map[string]fieldDoc
			if docType.Decl != nil && len(docType.Decl.Specs) > 0 {
				spec := docType.Decl.Specs[0]
				fieldMap = goDocForFields(spec)
			}

			fields := map[string]string{}
			for key, docField := range fieldMap {
				openApiKey := key

				info := NewTagInfo(docField.Tag)
				if info.HasName() {
					openApiKey = info.Name()
				}

				fields[openApiKey] = docField.Comment
			}

			return comment, fields, nil
		}
	}

	return "", nil, nil
}
