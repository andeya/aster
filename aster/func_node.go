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

// FuncType function type
type FuncType struct {
	*super
	*ast.FuncLit
	recv    *FuncField
	params  []*FuncField
	results []*FuncField
}

func (f *File) newFuncType(namePtr *string, doc *ast.CommentGroup,
	node *ast.FuncLit, recv *FuncField, params, results []*FuncField) *FuncType {
	ft := &FuncType{
		super:   f.newSuper(namePtr, Func, doc),
		FuncLit: node,
		recv:    recv,
		params:  params,
		results: results,
	}
	return ft
}

func (f *FuncType) funcNode() {}

// NumParam returns a function type's input parameter count.
func (f *FuncType) NumParam() int {
	return len(f.params)
}

// NumResult returns a function type's output parameter count.
func (f *FuncType) NumResult() int {
	return len(f.results)
}

// Param returns the type of a function type's i'th input parameter.
func (f *FuncType) Param(i int) (ff *FuncField, found bool) {
	if i < 0 || i >= len(f.params) {
		return
	}
	return f.params[i], true
}

// Result returns the type of a function type's i'th output parameter.
func (f *FuncType) Result(i int) (ff *FuncField, found bool) {
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
func (f *FuncType) IsVariadic() bool {
	return isVariadic(f.FuncLit.Type)
}

// Recv returns receiver (methods); or returns false (functions)
func (f *FuncType) Recv() (*FuncField, bool) {
	return f.recv, f.recv != nil
}
