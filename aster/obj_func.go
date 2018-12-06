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
	"fmt"
	"go/ast"
)

// funcObject function Declaration
type funcObject struct {
	*super
	node     ast.Node // *ast.FuncLit or *ast.FuncDecl
	isGlobal bool
	recv     *FuncField
	params   []*FuncField
	results  []*FuncField
}

// FuncField function params or results.
type FuncField struct {
	Name     string
	TypeName string // not contain `*`
}

var _ Object = (*funcObject)(nil)
var _ FuncObject = (*funcObject)(nil)

func (f *File) newFuncObject(namePtr *string, doc *ast.CommentGroup,
	node ast.Node, recv *FuncField, params, results []*FuncField) *funcObject {
	var isGlobal bool
	var objKind ast.ObjKind
	switch node.(type) {
	case *ast.FuncLit:
		objKind = ast.Var
		isGlobal = f.isGlobalTypOrFun(namePtr, node)
	case *ast.FuncDecl:
		objKind = ast.Fun
		isGlobal = true
	default:
		panic(fmt.Sprintf("want: *ast.FuncLit or *ast.FuncDecl, but got: %T", node))
	}
	ft := &funcObject{
		super:    f.newSuper(namePtr, objKind, Func, doc),
		node:     node,
		isGlobal: isGlobal,
		recv:     recv,
		params:   params,
		results:  results,
	}
	return ft
}

func (f *funcObject) funcObjectIdentify() {}

// objType returns the node that declares the object type.
func (f *funcObject) objType() ast.Node {
	return f.node
}

// IsGlobal returns whether the declaration is global.
func (f *funcObject) IsGlobal() bool {
	return f.isGlobal
}

// String returns the formated code block.
func (f *funcObject) String() string {
	s, err := f.file.FormatNode(f.objType())
	if err != nil {
		return fmt.Sprintf("// Formatting error: %s", err.Error())
	}
	if f.ObjKind() == ast.Fun {
		return s
	}
	s = "var " + f.Name() + " = " + s
	doc := f.Doc()
	if doc != "" {
		s = "// " + doc + s
	}
	return s
}

// NumParam returns a function type's input parameter count.
func (f *funcObject) NumParam() int {
	return len(f.params)
}

// NumResult returns a function type's output parameter count.
func (f *funcObject) NumResult() int {
	return len(f.results)
}

// Param returns the type of a function type's i'th input parameter.
func (f *funcObject) Param(i int) (ff *FuncField, found bool) {
	if i < 0 || i >= len(f.params) {
		return
	}
	return f.params[i], true
}

// Result returns the type of a function type's i'th output parameter.
func (f *funcObject) Result(i int) (ff *FuncField, found bool) {
	if i < 0 || i >= len(f.results) {
		return
	}
	return f.results[i], true
}

// IsVariadic reports whether a function type's final input parameter
// is a "..." parameter. If so, t.In(t.NumIn() - 1) returns the parameter's
// implicit actual type []T.
//
// For concreteness, if t represents func(x int, y ... float64), then
//
//	f.NumParam() == 2
//	f.Param(0) is the Type for "int"
//	f.Param(1) is the Type for "[]float64"
//	f.IsVariadic() == true
//
func (f *funcObject) IsVariadic() bool {
	switch t := f.node.(type) {
	case *ast.FuncLit:
		return isVariadic(t.Type)
	case *ast.FuncDecl:
		return isVariadic(t.Type)
	default:
		return false
	}
}

// Recv returns receiver (methods); or returns false (functions)
func (f *funcObject) Recv() (*FuncField, bool) {
	return f.recv, f.recv != nil
}
