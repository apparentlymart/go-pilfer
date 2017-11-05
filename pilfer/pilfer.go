package pilfer

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"io"

	"golang.org/x/tools/go/loader"
)

func Pilfer(srcPkg, typeName string, w io.Writer, pkgName string) error {
	prog, err := sourceProgram(srcPkg)
	if err != nil {
		return err
	}
	info := prog.Imported[srcPkg]

	ty := findTypeNameString(info, typeName)
	if ty == nil {
		return fmt.Errorf("package %s contains no type named %q", info.Pkg.Name(), typeName)
	}
	if !ty.IsNamed() {
		return fmt.Errorf("type %s in package %s is not a named type", typeName, info.Pkg.Name())
	}

	table := findInterestingTypes(ty, prog)
	names := table.NewNames()

	buf := bytes.Buffer{}
	fmt.Fprintf(&buf, "package %s\n\n", pkgName)

	for _, newName := range names {
		ty := table.TypeByNewName(newName)
		pkgPath := ty.Name.Pkg().Path()
		info := prog.Package(pkgPath)

		wrap := &ast.GenDecl{
			Tok: token.TYPE,
			Specs: []ast.Spec{
				ty.Spec,
			},
		}
		rewriteTypeIdents(wrap, info, table)
		format.Node(&buf, prog.Fset, wrap)
		buf.WriteString("\n\n")
	}

	fmted, err := format.Source(buf.Bytes())
	if err != nil {
		// should never happen because we should always generate valid input
		panic(err)
	}
	w.Write(fmted)

	return nil
}

func sourceProgram(srcPkg string) (*loader.Program, error) {
	var cfg loader.Config
	cfg.Import(srcPkg)
	return cfg.Load()
}

func findTypeNameString(info *loader.PackageInfo, typeName string) *takeType {
	var typeIdent *ast.Ident
	var typeSpec *ast.TypeSpec
	for _, file := range info.Files {
		for _, decl := range file.Decls {
			if gd, isGen := decl.(*ast.GenDecl); isGen {
				for _, spec := range gd.Specs {
					if ts, isType := spec.(*ast.TypeSpec); isType {
						if ts.Name.Name == typeName {
							typeIdent = ts.Name
							typeSpec = ts
						}
					}
				}
			}
		}
	}

	if typeIdent == nil {
		return nil
	}

	typeObj := info.Defs[typeIdent]
	if typeObj == nil {
		return nil
	}

	if typeName, isName := typeObj.(*types.TypeName); isName {
		tyTy := typeName.Type()
		return &takeType{
			Name:  typeName,
			Ident: typeIdent,
			Spec:  typeSpec,
			Type:  tyTy,
		}
	}

	return nil
}

func findTypeName(prog *loader.Program, name *types.TypeName) *takeType {
	if name == nil {
		return nil
	}
	pkg := name.Pkg()
	if pkg == nil {
		return nil
	}
	pkgPath := pkg.Path()
	typeName := name.Name()
	info := prog.Package(pkgPath)

	var typeIdent *ast.Ident
	var typeSpec *ast.TypeSpec
	for _, file := range info.Files {
		for _, decl := range file.Decls {
			if gd, isGen := decl.(*ast.GenDecl); isGen {
				for _, spec := range gd.Specs {
					if ts, isType := spec.(*ast.TypeSpec); isType {
						if ts.Name.Name == typeName {
							typeIdent = ts.Name
							typeSpec = ts
						}
					}
				}
			}
		}
	}

	if typeIdent == nil {
		return nil
	}

	typeObj := info.Defs[typeIdent]
	if typeObj == nil {
		return nil
	}

	if typeName, isName := typeObj.(*types.TypeName); isName {
		tyTy := typeName.Type()
		return &takeType{
			Name:  typeName,
			Ident: typeIdent,
			Spec:  typeSpec,
			Type:  tyTy,
		}
	}

	return nil
}

func findInterestingTypes(start *takeType, prog *loader.Program) typeTable {
	ret := newTypeTable()
	addInterestingTypes(start, ret, prog)
	return ret
}

func addInterestingTypes(start *takeType, table typeTable, prog *loader.Program) {
	table.Add(start)
	info := prog.Package(start.Name.Pkg().Path())
	astVisitor(func(node ast.Node) {
		ident, isIdent := node.(*ast.Ident)
		if !isIdent {
			return
		}

		obj := info.Uses[ident]
		if obj == nil {
			return
		}

		tn, isTypeName := obj.(*types.TypeName)
		if !isTypeName {
			return
		}

		if tn.Pkg() == nil {
			// Built-in types don't have packages, but we don't care about
			// them anyway.
			return
		}

		ty := findTypeName(prog, tn)
		if ty != nil && !table.Has(ty) {
			addInterestingTypes(ty, table, prog)
		}
	}).VisitAll(start.Spec)
}

func rewriteTypeIdents(start ast.Node, info *loader.PackageInfo, table typeTable) {
	astVisitor(func(node ast.Node) {
		switch tn := node.(type) {
		case *ast.TypeSpec:
			ident := tn.Name
			obj := info.Defs[ident]
			if obj == nil {
				return
			}
			if name, isName := obj.(*types.TypeName); isName {
				ty := table.TypeByName(name)
				if ty == nil {
					return
				}
				tn.Name = &ast.Ident{
					Name: ty.NewName,
				}
			}
		case *ast.Field:
			tn.Type = rewriteTypeExpr(tn.Type, info, table)
		case *ast.MapType:
			tn.Key = rewriteTypeExpr(tn.Key, info, table)
			tn.Value = rewriteTypeExpr(tn.Value, info, table)
		case *ast.ArrayType:
			tn.Elt = rewriteTypeExpr(tn.Elt, info, table)
		case *ast.ChanType:
			tn.Value = rewriteTypeExpr(tn.Value, info, table)
		case *ast.StarExpr:
			tn.X = rewriteTypeExpr(tn.X, info, table)
		}
	}).VisitAll(start)
}

func rewriteTypeExpr(expr ast.Expr, info *loader.PackageInfo, table typeTable) ast.Expr {
	switch tn := expr.(type) {

	case *ast.SelectorExpr:
		ident := tn.Sel
		obj := info.Uses[ident]
		if obj == nil {
			return expr
		}
		if name, isName := obj.(*types.TypeName); isName {
			ty := table.TypeByName(name)
			if ty == nil {
				return expr
			}
			return &ast.Ident{
				Name: ty.NewName,
			}
		}

	case *ast.Ident:
		obj := info.Uses[tn]
		if obj == nil {
			return expr
		}
		if name, isName := obj.(*types.TypeName); isName {
			ty := table.TypeByName(name)
			if ty == nil {
				return expr
			}
			return &ast.Ident{
				Name: ty.NewName,
			}
		}

	}

	return expr
}
