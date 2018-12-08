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

// An Facade describes a named language entity such as a package,
// constant, type, variable, function (incl. methods), or label.
// Facade interface implement all the objects.
//
type Facade struct {
	obj   types.Object
	pkg   *PackageInfo
	ident *ast.Ident
	doc   *ast.CommentGroup
}

func (p *PackageInfo) getFacade(ident *ast.Ident) (facade *Facade, idx int) {
	for idx, facade = range p.facades {
		if facade.ident == ident {
			return
		}
	}
	return nil, -1
}

func (p *PackageInfo) addFacade(ident *ast.Ident, obj types.Object) {
	p.facades = append(p.facades, &Facade{
		obj:   obj,
		pkg:   p,
		ident: ident,
		doc:   p.DocComment(ident),
	})
}

func (p *PackageInfo) removeFacade(ident *ast.Ident) {
	_, idx := p.getFacade(ident)
	if idx != -1 {
		p.facades = append(p.facades[:idx], p.facades[idx+1:]...)
	}
}

// Ident returns the indent.
func (fa *Facade) Ident() *ast.Ident {
	return fa.ident
}

// Object returns the types.Object.
func (fa *Facade) Object() types.Object {
	return fa.obj
}

// ObjKind returns what the Facade represents.
func (fa *Facade) ObjKind() ObjKind {
	return GetObjKind(fa.obj)
}

// TypKind returns what the Facade type represents.
func (fa *Facade) TypKind() TypKind {
	if fa.ObjKind() == Bad {
		return Invalid
	}
	return GetTypKind(fa.typ())
}

// Name returns the type's name within its package for a defined type.
// For other (non-defined) types it returns the empty string.
func (fa *Facade) Name() string {
	return fa.ident.Name
}

// Doc returns lead comment.
func (fa *Facade) Doc() string {
	return fa.doc.Text()
}

// CoverDoc covers lead comment if it exists.
func (fa *Facade) CoverDoc(text string) bool {
	if fa.doc == nil {
		return false
	}
	fa.doc.List = fa.doc.List[len(fa.doc.List)-1:]
	doc := fa.doc.List[0]
	doc.Text = text
	text = "// " + strings.Replace(fa.doc.Text(), "\n", "\n// ", -1)
	doc.Text = text[:len(text)-3]
	return true
}

// String previews the object formated code and comment.
func (fa *Facade) String() string {
	return fa.pkg.Preview(fa.ident)
}

// ---------------------------------- ObjKind != Bad (package or _=v) ----------------------------------

func (fa *Facade) typ() types.Type {
	return fa.obj.Type()
}

// Underlying returns the underlying type of a type.
func (fa *Facade) Underlying() types.Type {
	return fa.typ().Underlying()
}

// ---------------------------------- TypKind = Signature (function) ----------------------------------

func (fa *Facade) signature() *types.Signature {
	return fa.typ().(*types.Signature)
}

// IsMethod returns whether it is a method.
// NOTE: Panic, if TypKind != Signature
func (fa *Facade) IsMethod() bool {
	return fa.signature().Recv() != nil
}

// Params returns the parameters of signature s, or nil.
// NOTE: Panic, if TypKind != Signature
func (fa *Facade) Params() *types.Tuple {
	return fa.signature().Params()
}

// Recv returns the receiver of signature s (if a method), or nil if a
// function. It is ignored when comparing signatures for identity.
//
// For an abstract method, Recv returns the enclosing interface either
// as a *Named or an *Interface. Due to embedding, an interface may
// contain methods whose receiver type is a different interface.
// NOTE: Panic, if TypKind != Signature
func (fa *Facade) Recv() *types.Var {
	return fa.signature().Recv()
}

// Results returns the results of signature s, or nil.
// NOTE: Panic, if TypKind != Signature
func (fa *Facade) Results() *types.Tuple {
	return fa.signature().Results()
}

// Variadic reports whether the signature s is variadic.
// NOTE: Panic, if TypKind != Signature
func (fa *Facade) Variadic() bool {
	return fa.signature().Variadic()
}
