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
	"go/types"
	"strings"
)

// An Object describes a named language entity such as a package,
// constant, type, variable, function (incl. methods), or label.
// All objects implement the Object interface.
//
type Object interface {
	// Object returns the types.Object.
	Object() types.Object

	// ObjKind returns what the Object represents.
	ObjKind() ObjKind

	TypKind() TypKind

	// Name returns the type's name within its package for a defined type.
	// For other (non-defined) types it returns the empty string.
	Name() string

	// Doc returns lead comment.
	Doc() string

	// CoverDoc covers lead comment if it exists.
	CoverDoc(string) bool

	// String previews the object formated code and comment.
	String() string
}

type object struct {
	pkg   *PackageInfo
	ident *ast.Ident
	obj   types.Object
	doc   *ast.CommentGroup
}

func (p *PackageInfo) newObject(ident *ast.Ident, obj types.Object) *object {
	return &object{
		pkg:   p,
		ident: ident,
		obj:   obj,
		doc:   p.DocComment(ident),
	}
}

// Object returns the types.Object.
func (obj *object) Object() types.Object {
	return obj.obj
}

// ObjKind returns what the Object represents.
func (obj *object) ObjKind() ObjKind {
	switch obj.obj.(type) {
	case *types.PkgName:
		return Pkg
	case *types.Const:
		return Con
	case *types.TypeName:
		return Typ
	case *types.Var:
		return Var
	case *types.Func:
		return Fun
	case *types.Label:
		return Lbl
	case *types.Builtin:
		return Bui
	case *types.Nil:
		return Nil
	}
	return Bad
}

// TypKind returns what the Object type represents.
func (obj *object) TypKind() TypKind {
	switch obj.obj.Type().(type) {
	case *types.Basic:
		return Basic
	case *types.Array:
		return Array
	case *types.Slice:
		return Slice
	case *types.Struct:
		return Struct
	case *types.Pointer:
		return Pointer
	case *types.Tuple:
		return Tuple
	case *types.Signature:
		return Signature
	case *types.Interface:
		return Interface
	case *types.Map:
		return Map
	case *types.Chan:
		return Chan
	case *types.Named:
		return Named
	}
	return Invalid
}

// Name returns the type's name within its package for a defined type.
// For other (non-defined) types it returns the empty string.
func (obj *object) Name() string {
	return obj.ident.Name
}

// Doc returns lead comment.
func (obj *object) Doc() string {
	return obj.doc.Text()
}

// CoverDoc covers lead comment if it exists.
func (obj *object) CoverDoc(text string) bool {
	if obj.doc == nil {
		return false
	}
	obj.doc.List = obj.doc.List[len(obj.doc.List)-1:]
	doc := obj.doc.List[0]
	doc.Text = text
	text = "// " + strings.Replace(obj.doc.Text(), "\n", "\n// ", -1)
	doc.Text = text[:len(text)-3]
	return true
}

// String previews the object formated code and comment.
func (obj *object) String() string {
	return obj.pkg.PreviewObject(obj.ident)
}

// A Func represents a declared function, concrete method, or abstract
// (interface) method. Its Type() is always a *Signature.
// An abstract method may belong to many interfaces due to embedding.
type Func struct {
	*object
	typesFunc *types.Func
}

func (p *PackageInfo) addFunc(ident *ast.Ident, fn *types.Func) {
	fun := &Func{
		object:    p.newObject(ident, fn),
		typesFunc: fn,
	}
	p.objects[ident.Pos()] = fun
}

func (f *Func) IsMethod() bool {
	return f.typesFunc.Type().(*types.Signature).Recv() != nil
}
