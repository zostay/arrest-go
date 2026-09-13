package compilecost

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

// The rule the DSL packages keep, so that naming one of their types does not
// cost the importing package libopenapi's instantiation cascade (see
// golang/go#70511 and issue #101):
//
//  1. No exported type may reach a libopenapi type — through a field, an
//     embedded type, a type argument, or the signature of any method,
//     exported or not.
//  2. Any function whose signature or body mentions a libopenapi type must be
//     marked //go:noinline, so that no importer ever reads its body.
//
// "Libopenapi type" includes any type declared in the package that reaches
// one, such as the unexported state structs the DSL types point at.

// taintedPrefixes are the module paths whose types carry the cascade.
var taintedPrefixes = []string{
	"github.com/pb33f/",
	"go.yaml.in/",
	"gopkg.in/yaml",
}

// Opaque type-checks the package in dir and returns one line per violation
// of the rule above, sorted, or nothing when the package keeps it.
func Opaque(dir string) ([]string, error) {
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports,
		Dir: dir,
	}, ".")
	if err != nil {
		return nil, err
	}
	if len(pkgs) != 1 {
		return nil, fmt.Errorf("expected one package in %s, got %d", dir, len(pkgs))
	}
	pkg := pkgs[0]
	if len(pkg.Errors) > 0 {
		return nil, fmt.Errorf("loading %s: %v", pkg.PkgPath, pkg.Errors[0])
	}

	c := &checker{pkg: pkg, memo: map[types.Type]string{}}
	var violations []string

	// Rule 1: exported types.
	scope := pkg.Types.Scope()
	for _, name := range scope.Names() {
		tn, ok := scope.Lookup(name).(*types.TypeName)
		if !ok || !tn.Exported() || tn.IsAlias() {
			continue
		}
		if why := c.taint(tn.Type()); why != "" {
			violations = append(violations, fmt.Sprintf("%s: exported type %s reaches libopenapi: %s",
				pkg.Fset.Position(tn.Pos()), name, why))
		}
	}

	// Rule 2: functions without //go:noinline.
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || isNoinline(fd) {
				continue
			}
			obj, ok := pkg.TypesInfo.Defs[fd.Name].(*types.Func)
			if !ok {
				continue
			}
			if why := c.taint(obj.Type()); why != "" {
				violations = append(violations, fmt.Sprintf("%s: %s is not //go:noinline but its signature reaches libopenapi: %s",
					pkg.Fset.Position(fd.Pos()), funcName(fd), why))
				continue
			}
			if why, at := c.bodyTaint(fd); why != "" {
				violations = append(violations, fmt.Sprintf("%s: %s is not //go:noinline but its body reaches libopenapi at %s: %s",
					pkg.Fset.Position(fd.Pos()), funcName(fd), pkg.Fset.Position(at), why))
			}
		}
	}

	sort.Strings(violations)
	return violations, nil
}

type checker struct {
	pkg  *packages.Package
	memo map[types.Type]string
}

// taint reports how t reaches a libopenapi type, or "" if it does not.
func (c *checker) taint(t types.Type) string {
	return c.taintPath(t, nil)
}

func (c *checker) taintPath(t types.Type, seen map[types.Type]bool) string {
	if why, ok := c.memo[t]; ok {
		return why
	}
	if seen == nil {
		seen = map[types.Type]bool{}
	}
	if seen[t] {
		return ""
	}
	seen[t] = true

	why := c.taintOf(t, seen)
	c.memo[t] = why
	return why
}

func (c *checker) taintOf(t types.Type, seen map[types.Type]bool) string {
	switch t := t.(type) {
	case *types.Alias:
		return c.taintPath(types.Unalias(t), seen)
	case *types.Named:
		if obj := t.Obj(); obj.Pkg() != nil {
			if isTaintedPath(obj.Pkg().Path()) {
				return types.TypeString(t, nil)
			}
			// A type from some other package is opaque to us; only our own
			// and libopenapi's are followed.
			if obj.Pkg().Path() != c.pkg.PkgPath {
				return ""
			}
		}
		if args := t.TypeArgs(); args != nil {
			for i := 0; i < args.Len(); i++ {
				if why := c.taintPath(args.At(i), seen); why != "" {
					return why
				}
			}
		}
		if why := c.taintPath(t.Underlying(), seen); why != "" {
			return t.Obj().Name() + " -> " + why
		}
		for i := 0; i < t.NumMethods(); i++ {
			m := t.Method(i)
			if why := c.taintPath(m.Type(), seen); why != "" {
				return fmt.Sprintf("%s.%s -> %s", t.Obj().Name(), m.Name(), why)
			}
		}
		return ""
	case *types.Pointer:
		return c.taintPath(t.Elem(), seen)
	case *types.Slice:
		return c.taintPath(t.Elem(), seen)
	case *types.Array:
		return c.taintPath(t.Elem(), seen)
	case *types.Chan:
		return c.taintPath(t.Elem(), seen)
	case *types.Map:
		if why := c.taintPath(t.Key(), seen); why != "" {
			return why
		}
		return c.taintPath(t.Elem(), seen)
	case *types.Struct:
		for i := 0; i < t.NumFields(); i++ {
			f := t.Field(i)
			if why := c.taintPath(f.Type(), seen); why != "" {
				return "field " + f.Name() + " -> " + why
			}
		}
		return ""
	case *types.Signature:
		if r := t.Recv(); r != nil {
			if why := c.taintPath(r.Type(), seen); why != "" {
				return "receiver -> " + why
			}
		}
		if why := c.taintPath(t.Params(), seen); why != "" {
			return why
		}
		return c.taintPath(t.Results(), seen)
	case *types.Tuple:
		for i := 0; i < t.Len(); i++ {
			if why := c.taintPath(t.At(i).Type(), seen); why != "" {
				return why
			}
		}
		return ""
	case *types.Interface:
		for i := 0; i < t.NumMethods(); i++ {
			m := t.Method(i)
			if why := c.taintPath(m.Type(), seen); why != "" {
				return "method " + m.Name() + " -> " + why
			}
		}
		return ""
	}
	// Basic types, type parameters, unions: nothing to reach.
	return ""
}

// bodyTaint finds the first expression in the function body whose type
// reaches libopenapi.
func (c *checker) bodyTaint(fd *ast.FuncDecl) (why string, at token.Pos) {
	if fd.Body == nil {
		return "", token.NoPos
	}
	ast.Inspect(fd.Body, func(n ast.Node) bool {
		if why != "" {
			return false
		}
		expr, ok := n.(ast.Expr)
		if !ok {
			return true
		}
		t := c.pkg.TypesInfo.TypeOf(expr)
		if t == nil {
			return true
		}
		if w := c.taint(t); w != "" {
			why, at = w, expr.Pos()
			return false
		}
		return true
	})
	return why, at
}

func isTaintedPath(path string) bool {
	for _, p := range taintedPrefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

func isNoinline(fd *ast.FuncDecl) bool {
	if fd.Doc == nil {
		return false
	}
	for _, c := range fd.Doc.List {
		if strings.TrimSpace(c.Text) == "//go:noinline" {
			return true
		}
	}
	return false
}

func funcName(fd *ast.FuncDecl) string {
	if fd.Recv == nil || len(fd.Recv.List) == 0 {
		return fd.Name.Name
	}
	recv := fd.Recv.List[0].Type
	if star, ok := recv.(*ast.StarExpr); ok {
		recv = star.X
	}
	if idx, ok := recv.(*ast.IndexExpr); ok {
		recv = idx.X
	}
	if id, ok := recv.(*ast.Ident); ok {
		return id.Name + "." + fd.Name.Name
	}
	return fd.Name.Name
}
