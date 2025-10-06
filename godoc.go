package arrest

import (
	"fmt"
	"go/ast"
	"go/doc"
	"reflect"
	"strings"
	"sync"

	"golang.org/x/tools/go/packages"
)

type fieldDoc struct {
	Name    string
	Comment string
	Tag     reflect.StructTag
}

var (
	packageCache = make(map[string]*doc.Package)
	cacheMutex   sync.RWMutex
)

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

func getPackageDoc(pkgPath string) (*doc.Package, error) {
	cacheMutex.RLock()
	if cached, exists := packageCache[pkgPath]; exists {
		cacheMutex.RUnlock()
		return cached, nil
	}
	cacheMutex.RUnlock()

	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	// Double-check after acquiring write lock
	if cached, exists := packageCache[pkgPath]; exists {
		return cached, nil
	}

	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedFiles,
	}, pkgPath)
	if err != nil {
		return nil, err
	}

	if len(pkgs) == 0 {
		packageCache[pkgPath] = nil
		return nil, nil
	}

	pkg := pkgs[0]
	if pkg.Fset == nil || pkg.Syntax == nil {
		packageCache[pkgPath] = nil
		return nil, nil
	}

	docPkg, err := doc.NewFromFiles(pkg.Fset, pkg.Syntax, pkgPath)
	if err != nil {
		return nil, err
	}

	packageCache[pkgPath] = docPkg
	return docPkg, nil
}

func GoDocForStruct(t reflect.Type) (string, map[string]string, error) {
	if t.Kind() != reflect.Struct {
		return "", nil, fmt.Errorf("expected a struct type, got %s", t.Kind())
	}

	if t.PkgPath() == "" {
		return "", nil, nil
	}

	docPkg, err := getPackageDoc(t.PkgPath())
	if err != nil {
		return "", nil, err
	}
	if docPkg == nil {
		return "", nil, nil
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

				// Rewrite commend to use the openapi name rather than the go name
				ps := strings.SplitN(docField.Comment, " ", 2)
				newComment := docField.Comment
				if len(ps) == 2 {
					firstWord, theRest := ps[0], ps[1]
					if firstWord == key {
						newComment = strings.Join([]string{openApiKey, theRest}, " ")
					}
				}

				fields[openApiKey] = newComment
			}

			return strings.TrimSpace(comment), fields, nil
		}
	}

	return "", nil, nil
}

// GoDocForType extracts godoc comments for any named type (struct, type alias, etc.)
func GoDocForType(t reflect.Type) string {
	if t.PkgPath() == "" || t.Name() == "" {
		return ""
	}

	// First try the existing GoDocForStruct for struct types
	if t.Kind() == reflect.Struct {
		if comment, _, err := GoDocForStruct(t); err == nil && comment != "" {
			return comment
		}
	}

	// For non-struct types, we need to manually extract from package doc
	docPkg, err := getPackageDoc(t.PkgPath())
	if err != nil || docPkg == nil {
		return ""
	}

	// Look for the type in the package documentation
	for _, docType := range docPkg.Types {
		if docType.Name == t.Name() {
			return strings.TrimSpace(docType.Doc)
		}
	}

	return ""
}
