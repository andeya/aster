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
)

func (p *Package) collectTypes() {
	for _, f := range p.Files {
		f.collectTypes(false)
	}
	// Waiting for types ready to do method association
	for _, f := range p.Files {
		f.collectMethods()
		f.addFuncParamAndResult()
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
		f.addFuncParamAndResult()
	}
}

func (f *File) collectFuncs() {
	collectFuncs := func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncLit:
			t := newFuncType(x, "", "", nil)
			f.Types[t.String()] = t
		case *ast.FuncDecl:
			if x.Recv != nil {
				return true
			}
			t := newFuncType(x, x.Name.Name, f.PkgName, x.Doc)
			f.Types[t.String()] = t
		}
		return true
	}
	ast.Inspect(f.File, collectFuncs)
}

func (f *File) addFuncParamAndResult() {
	// TODO
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
			Params:     []Type{},
			Result:     []Type{},
			IsVariadic: isVariadic(x.Type),
		}
		r.addMethods(m)
		return true
	}
	ast.Inspect(f.File, collectMethods)
}
