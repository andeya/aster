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
	"go/ast"
	"go/token"
	"strings"
)

// Inspect traverses objects in the module.
func (m *Module) Inspect(fn func(Object) bool) {
	for _, p := range m.Packages {
		p.Inspect(fn)
	}
}

// Fetch traversing through the current module, fetches node if fn returns true.
func (m *Module) Fetch(fn func(Object) bool) (objects []Object) {
	for _, p := range m.Packages {
		p.Inspect(func(obj Object) bool {
			if fn(obj) {
				objects = append(objects, obj)
			}
			return true
		})
	}
	return objects
}

// Module returns module object if exist.
func (p *Package) Module() (*Module, bool) {
	return p.module, p.module != nil
}

// Inspect traverses objects in the package.
func (p *Package) Inspect(fn func(Object) bool) {
	for _, f := range p.Files {
		f.Inspect(fn)
	}
}

// Fetch traversing through the current package, fetches node if fn returns true.
func (p *Package) Fetch(fn func(Object) bool) (objects []Object) {
	p.Inspect(func(obj Object) bool {
		if fn(obj) {
			objects = append(objects, obj)
		}
		return false
	})
	return objects
}

// LookupType lookups TypeObject by type name in the current package.
func (p *Package) LookupType(name string) (t TypeObject, found bool) {
	fn, ok := typeObjectFuncByNameInPkg(name)
	if !ok {
		return
	}
	var objects []Object
	for _, v := range p.Files {
		objects = v.Fetch(fn)
		if len(objects) > 0 {
			return objects[0].(TypeObject), true
		}
	}
	return
}

// LookupPureFunc lookups FuncObject by name in the current package.
// NOTE: lookup FuncObject, but not method function.
func (p *Package) LookupPureFunc(name string) (f FuncObject, found bool) {
	fn, ok := pureFuncObjectFuncByNameInPkg(name)
	if !ok {
		return
	}
	var objects []Object
	for _, v := range p.Files {
		objects = v.Fetch(fn)
		if len(objects) > 0 {
			return objects[0].(FuncObject), true
		}
	}
	return
}

// Package returns package object if exist.
func (f *File) Package() (*Package, bool) {
	return f.pkg, f.pkg != nil
}

// Inspect traverses objects in the file.
func (f *File) Inspect(fn func(Object) bool) {
	for _, obj := range f.Objects {
		if !fn(obj) {
			return
		}
	}
}

// Fetch traversing through the current file, fetches node if fn returns true.
func (f *File) Fetch(fn func(Object) bool) (objects []Object) {
	f.Inspect(func(obj Object) bool {
		if fn(obj) {
			objects = append(objects, obj)
		}
		return true
	})
	return objects
}

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
		if p, ok := mod.Packages[imp.Name]; ok {
			pkgs = append(pkgs, p)
			found = true
		}
	}
	return
}

// LookupType lookups TypeObject by type name in the current file.
func (f *File) LookupType(name string) (t TypeObject, found bool) {
	fn, ok := typeObjectFuncByNameInPkg(name)
	if !ok {
		return
	}
	objects := f.Fetch(fn)
	if len(objects) > 0 {
		return objects[0].(TypeObject), true
	}
	return
}

// LookupTypeInPkg lookups TypeObject by type name in the current package.
func (f *File) LookupTypeInPkg(name string) (t TypeObject, found bool) {
	p, ok := f.Package()
	if ok {
		return p.LookupType(name)
	}
	return f.LookupType(name)
}

// LookupTypeInMod lookup Type by type name in the current module.
func (f *File) LookupTypeInMod(name string) (t TypeObject, found bool) {
	p, ok := f.Package()
	if ok {
		t, found = p.LookupType(name)
	} else {
		t, found = f.LookupType(name)
	}
	if found {
		return
	}
	name = strings.TrimLeft(name, "*")
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
		t, found = p.LookupType(a[1])
		if found {
			return
		}
	}
	return
}

// LookupPureFunc lookups FuncObject by name in the current file.
// NOTE: lookup FuncObject, but not method function.
func (f *File) LookupPureFunc(name string) (pf FuncObject, found bool) {
	fn, ok := pureFuncObjectFuncByNameInPkg(name)
	if !ok {
		return
	}
	objects := f.Fetch(fn)
	if len(objects) > 0 {
		return objects[0].(FuncObject), true
	}
	return
}

// LookupPureFuncInPkg lookups FuncObject by name in the current package.
// NOTE: lookup FuncObject, but not method function.
func (f *File) LookupPureFuncInPkg(name string) (pf FuncObject, found bool) {
	p, ok := f.Package()
	if ok {
		return p.LookupPureFunc(name)
	}
	return f.LookupPureFunc(name)
}

// LookupPureFuncInMod lookup FuncObject by name in the current module.
// NOTE: lookup FuncObject, but not method function.
func (f *File) LookupPureFuncInMod(name string) (pf FuncObject, found bool) {
	p, ok := f.Package()
	if ok {
		pf, found = p.LookupPureFunc(name)
	} else {
		pf, found = f.LookupPureFunc(name)
	}
	if found {
		return
	}
	name = strings.TrimLeft(name, "*")
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
		pf, found = p.LookupPureFunc(a[1])
		if found {
			return
		}
	}
	return
}

func (p *Package) collectObjects() {
	for _, f := range p.Files {
		f.collectObjects(false)
	}
	// Waiting for types ready to do method association
	for _, f := range p.Files {
		f.bindMethods()
	}
}

// Use the method if no other file in the same package,
// otherwise use *Package.collectObjects()
func (f *File) collectObjects(singleParsing bool) {
	f.Objects = make(map[token.Pos]Object)
	f.collectTypesOtherThanStruct()
	f.collectFuncs()
	f.collectStructs()
	f.setStructFields()
	if singleParsing {
		f.bindMethods()
	}
}

func (f *File) collectFuncs() {
	collectFuncs := func(n ast.Node) bool {
		var t *funcObject
		switch x := n.(type) {
		case *ast.FuncDecl:
			var recv *FuncField
			if recvs := f.expandFuncFields(x.Recv); len(recvs) > 0 {
				recv = recvs[0]
			}
			t = f.newFuncObject(
				&x.Name.Name,
				x.Doc,
				x,
				recv,
				f.expandFuncFields(x.Type.Params),
				f.expandFuncFields(x.Type.Results),
			)
		default:
			return true
		}
		f.Objects[t.objType().Pos()] = t
		return true
	}
	ast.Inspect(f.File, collectFuncs)

	// recover value functions
	f.collectValueSpecs(func(n *ast.ValueSpec, doc *ast.CommentGroup) {
		if n.Doc != nil {
			doc = n.Doc
		}
		for k, v := range n.Values {
			fl, ok := v.(*ast.FuncLit)
			if !ok {
				continue
			}
			t := f.newFuncObject(
				&n.Names[k].Name,
				doc,
				fl,
				nil,
				f.expandFuncFields(fl.Type.Params),
				f.expandFuncFields(fl.Type.Results),
			)
			f.Objects[t.objType().Pos()] = t
		}
	})
}

func (f *File) collectValueSpecs(fn func(*ast.ValueSpec, *ast.CommentGroup)) {
	ast.Inspect(f.File, func(n ast.Node) bool {
		if decl, ok := n.(*ast.GenDecl); ok {
			doc := decl.Doc
			for _, spec := range decl.Specs {
				if td, ok := spec.(*ast.ValueSpec); ok {
					if td.Doc != nil {
						doc = td.Doc
					}
					fn(td, doc)
				}
			}
		}
		return true
	})
}

func (f *File) collectTypeSpecs(fn func(*ast.TypeSpec, *ast.CommentGroup)) {
	ast.Inspect(f.File, func(n ast.Node) bool {
		if decl, ok := n.(*ast.GenDecl); ok {
			doc := decl.Doc
			for _, spec := range decl.Specs {
				if td, ok := spec.(*ast.TypeSpec); ok {
					if td.Doc != nil {
						doc = td.Doc
					}
					fn(td, doc)
				}
			}
		}
		return true
	})
}

func (f *File) collectTypesOtherThanStruct() {
	f.collectTypeSpecs(func(node *ast.TypeSpec, doc *ast.CommentGroup) {
		namePtr := &node.Name.Name
		var t Object
		if getElem(node.Type) != node.Type {
			t = f.newAliasType(namePtr, doc, node.Assign, node.Type)
		} else {
			switch x := node.Type.(type) {
			case *ast.SelectorExpr:
				t = f.newAliasType(namePtr, doc, node.Assign, x)

			case *ast.Ident:
				t = f.newBasicOrAliasType(namePtr, doc, node.Assign, x)

			case *ast.ChanType:
				t = f.newChanType(namePtr, doc, node.Assign, x)

			case *ast.ArrayType:
				t = f.newListType(namePtr, doc, node.Assign, x)

			case *ast.MapType:
				t = f.newMapType(namePtr, doc, node.Assign, x)

			case *ast.InterfaceType:
				t = f.newInterfaceType(namePtr, doc, node.Assign, x)

			default:
				return
			}
		}
		f.Objects[t.objType().Pos()] = t
	})
}

// collectStructs collects and maps structType objects to their positions
func (f *File) collectStructs() {
	collectStructs := func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CompositeLit:
			t, ok := x.Type.(*ast.StructType)
			if !ok {
				return true
			}
			st := f.newStructType(nil, ast.Var, nil, -1, t)
			f.Objects[st.objType().Pos()] = st
		case *ast.GenDecl:
			for _, spec := range x.Specs {
				var assign = token.NoPos
				var t ast.Expr
				var structName *string
				var doc = x.Doc
				var objKind = ast.Typ
				switch y := spec.(type) {
				case *ast.TypeSpec:
					if y.Type == nil {
						continue
					}
					assign = y.Assign
					structName = &y.Name.Name
					t = y.Type
					if y.Doc != nil {
						doc = y.Doc
					}
				case *ast.ValueSpec:
					assign = -1
					structName = &y.Names[0].Name
					t = y.Type
					if y.Doc != nil {
						doc = y.Doc
					}
					objKind = ast.Var
				}
				z, ok := t.(*ast.StructType)
				if !ok {
					continue
				}
				st := f.newStructType(structName, objKind, doc, assign, z)
				f.Objects[st.objType().Pos()] = st
			}
		}
		return true
	}
	ast.Inspect(f.File, collectStructs)
}

func (f *File) setStructFields() {
	for _, t := range f.Objects {
		s, ok := t.(*structType)
		if !ok {
			continue
		}
		s.setFields()
	}
}

func (f *File) bindMethods() {
	for _, m := range f.Objects {
		fb, ok := m.(FuncObject)
		if !ok {
			continue
		}
		recv, found := fb.Recv()
		if !found {
			continue
		}
		t, found := f.LookupTypeInPkg(recv.TypeName)
		if !found {
			continue
		}
		t.addMethod(fb)
		break
	}
}

func (f *File) expandFuncFields(fieldList *ast.FieldList) (fields []*FuncField) {
	if fieldList != nil {
		for _, g := range fieldList.List {
			typeName := f.TryFormatNode(g.Type)
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

func expandFields(fieldList *ast.FieldList) {
	if fieldList == nil {
		return
	}
	var list = make([]*ast.Field, 0, fieldList.NumFields())
	for _, field := range fieldList.List {
		list = append(list, field)
		if len(field.Names) > 1 {
			for _, name := range field.Names[1:] {
				list = append(list, &ast.Field{
					Names: []*ast.Ident{cloneIdent(name)},
					Type:  field.Type,
					Tag:   cloneBasicLit(field.Tag),
				})
			}
			field.Names = field.Names[:1]
		}
	}
	fieldList.List = list
}

func typeObjectFuncByNameInPkg(name string) (func(Object) bool, bool) {
	if strings.Contains(name, ".") {
		return nil, false
	}
	name = strings.TrimLeft(name, "*")
	return func(obj Object) bool {
		return obj.ObjKind() == ast.Typ && obj.Name() == name
	}, true
}

// NOTE: lookup FuncObject, but not method function.
func pureFuncObjectFuncByNameInPkg(name string) (func(Object) bool, bool) {
	if strings.Contains(name, ".") {
		return nil, false
	}
	name = strings.TrimLeft(name, "*")
	return func(obj Object) bool {
		return obj.Name() == name && IsPureFunc(obj)
	}, true
}

// NOTE: lookup method FuncObject.
func methodNodeFuncByNameInPkg(name string) (func(Object) bool, bool) {
	if strings.Contains(name, ".") {
		return nil, false
	}
	name = strings.TrimLeft(name, "*")
	return func(obj Object) bool {
		return obj.Name() == name && IsMethod(obj)
	}, true
}

func getElem(e ast.Expr) ast.Expr {
	for {
		s, ok := e.(*ast.StarExpr)
		if ok {
			e = s.X
		} else {
			return e
		}
	}
}

func cloneIdent(i *ast.Ident) *ast.Ident {
	return &ast.Ident{
		Name: i.Name,
		Obj:  i.Obj,
	}
}

func cloneBasicLit(b *ast.BasicLit) *ast.BasicLit {
	if b == nil {
		return nil
	}
	return &ast.BasicLit{
		Kind:  b.Kind,
		Value: b.Value,
	}
}

// func cloneCommentGroup(c *ast.CommentGroup) *ast.CommentGroup {
// 	if c == nil {
// 		return nil
// 	}
// 	n := new(ast.CommentGroup)
// 	for _, v := range c.List {
// 		n.List = append(n.List, &ast.Comment{
// 			Text: v.Text,
// 		})
// 	}
// 	return n
// }
