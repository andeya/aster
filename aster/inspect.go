// Copyright 2018 henrylee2cn. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aster

import (
	"bytes"
	"go/ast"
	"go/format"
	"strings"
)

// LookupImports lookups the import info by package name.
func (f *File) LookupImports(currPkgName string) (imports []*Import, found bool) {
	for _, imp := range f.Imports {
		if imp.Name == currPkgName {
			imports = append(imports, imp)
			found = true
		}
	}
	return
}

// LookupPackages lookups the package object by package name.
// NOTE: Only lookup the parsed module.
func (f *File) LookupPackages(currPkgName string) (pkgs []*Package, found bool) {
	if f.pkg == nil || f.pkg.module == nil {
		return
	}
	imps, found := f.LookupImports(currPkgName)
	if !found {
		return
	}
	mod := f.pkg.module
	for _, imp := range imps {
		if p, ok := mod.Pkgs[imp.Name]; ok {
			pkgs = append(pkgs, p)
			found = true
		}
	}
	return
}

// LookupType lookup Type by type name.
func (f *File) LookupType(name string) (t Type, found bool) {
	name = strings.TrimLeft(name, "*")
	// May be basic type?
	t, found = getBasicType(name)
	if found {
		return
	}
	// May be in the current package?
	if !strings.Contains(name, ".") {
		if f.pkg == nil {
			t, found = f.Types[name]
			if found {
				return
			}
		} else {
			for _, v := range f.pkg.Files {
				t, found = v.Types[name]
				if found {
					return
				}
			}
		}
	}
	// May be in the other module packages?
	a := strings.SplitN(name, ".", 2)
	if len(a) == 1 {
		a = []string{".", name}
	}
	pkgs, ok := f.LookupPackages(a[0])
	if !ok {
		return
	}
	for _, p := range pkgs {
		for _, v := range p.Files {
			t, found = v.Types[a[1]]
			if found {
				return
			}
		}
	}
	return
}

func (p *Package) collectTypes() {
	for _, f := range p.Files {
		f.collectTypes(false)
	}
	// Waiting for types ready to do method association
	for _, f := range p.Files {
		f.collectMethods()
	}
}

// Use the method if no other file in the same package,
// otherwise use *Package.collectTypes()
func (f *File) collectTypes(collectMethods bool) {
	f.Types = make(map[string]Type)
	f.collectFuncs()
	f.collectStructs()
	if collectMethods {
		f.collectMethods()
	}
}

func (f *File) collectFuncs() {
	collectFuncs := func(n ast.Node) bool {
		var t *FuncType
		var funcType *ast.FuncType
		switch x := n.(type) {
		case *ast.FuncLit:
			funcType = x.Type
			t = newFuncType(x, "", "", nil)
		case *ast.FuncDecl:
			if x.Recv != nil {
				return true
			}
			funcType = x.Type
			t = newFuncType(x, x.Name.Name, f.PkgName, x.Doc)
		default:
			return true
		}
		t.params = f.expandFuncFields(funcType.Params)
		t.results = f.expandFuncFields(funcType.Results)
		f.Types[t.String()] = t
		return true
	}
	ast.Inspect(f.File, collectFuncs)
}

// func collectDecl(f *File) (decls []ast.Decl) {
// 	ast.Inspect(f.File, func(n ast.Node) bool {
// 		decl, ok := n.(ast.Decl)
// 		if ok {
// 			decls = append(decls, decl)
// 		}
// 		return true
// 	})
// 	return
// }

// collectStructs collects and maps structType nodes to their positions
func (f *File) collectStructs() {
	collectStructs := func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CompositeLit:
			t, ok := x.Type.(*ast.StructType)
			if !ok {
				return true
			}
			st := newStructType(t, "", "", nil)
			f.Types[st.String()] = st
		case *ast.GenDecl:
			var declDoc *ast.CommentGroup
			if len(x.Specs) == 1 {
				declDoc = x.Doc
			}
			for _, spec := range x.Specs {
				var t ast.Expr
				var structName string
				var doc = declDoc
				switch x := spec.(type) {
				case *ast.TypeSpec:
					if x.Type == nil {
						continue
					}
					structName = x.Name.Name
					t = x.Type
				case *ast.ValueSpec:
					structName = x.Names[0].Name
					t = x.Type
				}
				x, ok := t.(*ast.StructType)
				if !ok {
					continue
				}
				st := newStructType(x, structName, f.PkgName, doc)
				f.Types[st.String()] = st
			}
		}
		return true
	}
	ast.Inspect(f.File, collectStructs)
}

func (f *File) collectMethods() {
	collectMethods := func(n ast.Node) bool {
		x, ok := n.(*ast.FuncDecl)
		if !ok || x.Recv == nil || len(x.Recv.List) == 0 {
			return true
		}
		recvTypeName := x.Recv.List[0].Type.(*ast.StarExpr).X.(*ast.Ident).Name
		r, ok := f.LookupType(recvTypeName)
		if !ok {
			return true
		}
		m := &Method{
			FuncDecl:   x,
			Recv:       r,
			Name:       x.Name.Name,
			Doc:        x.Doc,
			Params:     f.expandFuncFields(x.Type.Params),
			Results:    f.expandFuncFields(x.Type.Results),
			IsVariadic: isVariadic(x.Type),
		}
		r.addMethods(m)
		return true
	}
	ast.Inspect(f.File, collectMethods)
}

func (f *File) expandFuncFields(fieldList *ast.FieldList) (fields []*FuncField) {
	if fieldList != nil {
		for _, g := range fieldList.List {
			typeName := f.tryFormat(g.Type)
			m := len(g.Names)
			if m == 0 {
				fields = append(fields, &FuncField{
					TypeName: typeName,
				})
			} else {
				for _, name := range g.Names {
					fields = append(fields, &FuncField{
						Name:     name.Name,
						TypeName: typeName,
					})
				}
			}
		}
	}
	return
}

func (f *File) format(node ast.Node) (code string, err error) {
	var dst bytes.Buffer
	err = format.Node(&dst, f.FileSet, node)
	if err != nil {
		return
	}
	return dst.String(), nil
}

func (f *File) tryFormat(node ast.Node, defaultValue ...string) string {
	code, err := f.format(node)
	if err != nil && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return code
}
