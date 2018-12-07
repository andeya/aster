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

// An Aster describes a named language entity such as a package,
// constant, type, variable, function (incl. methods), or label.
// All objects implement the Aster interface.
//
type Aster struct {
	obj   types.Object
	pkg   *PackageInfo
	ident *ast.Ident
	doc   *ast.CommentGroup
}

func (p *PackageInfo) getAster(ident *ast.Ident) (aster *Aster, idx int) {
	for idx, aster = range p.asters {
		if aster.ident == ident {
			return
		}
	}
	return nil, -1
}

func (p *PackageInfo) addAster(ident *ast.Ident, obj types.Object) {
	p.asters = append(p.asters, &Aster{
		obj:   obj,
		pkg:   p,
		ident: ident,
		doc:   p.DocComment(ident),
	})
}

func (p *PackageInfo) removeAster(ident *ast.Ident) {
	_, idx := p.getAster(ident)
	if idx != -1 {
		p.asters = append(p.asters[:idx], p.asters[idx+1:]...)
	}
}

// Ident returns the indent.
func (a *Aster) Ident() *ast.Ident {
	return a.ident
}

// Object returns the types.Object.
func (a *Aster) Object() types.Object {
	return a.obj
}

// ObjKind returns what the Aster represents.
func (a *Aster) ObjKind() ObjKind {
	return GetObjKind(a.obj)
}

// TypKind returns what the Aster type represents.
func (a *Aster) TypKind() TypKind {
	if a.ObjKind() == Bad {
		return Invalid
	}
	return GetTypKind(a.typ())
}

// Name returns the type's name within its package for a defined type.
// For other (non-defined) types it returns the empty string.
func (a *Aster) Name() string {
	return a.ident.Name
}

// Doc returns lead comment.
func (a *Aster) Doc() string {
	return a.doc.Text()
}

// CoverDoc covers lead comment if it exists.
func (a *Aster) CoverDoc(text string) bool {
	if a.doc == nil {
		return false
	}
	a.doc.List = a.doc.List[len(a.doc.List)-1:]
	doc := a.doc.List[0]
	doc.Text = text
	text = "// " + strings.Replace(a.doc.Text(), "\n", "\n// ", -1)
	doc.Text = text[:len(text)-3]
	return true
}

// String previews the object formated code and comment.
func (a *Aster) String() string {
	return a.pkg.Preview(a.ident)
}

// ---------------------------------- ObjKind != Bad (package or _=v) ----------------------------------

func (a *Aster) typ() types.Type {
	return a.obj.Type()
}

// Underlying returns the underlying type of a type.
func (a *Aster) Underlying() types.Type {
	return a.typ().Underlying()
}

// ---------------------------------- TypKind = Signature (function) ----------------------------------

func (a *Aster) signature() *types.Signature {
	return a.typ().(*types.Signature)
}

// IsMethod returns whether it is a method.
// NOTE: Panic, if TypKind != Signature
func (a *Aster) IsMethod() bool {
	return a.signature().Recv() != nil
}

// Params returns the parameters of signature s, or nil.
// NOTE: Panic, if TypKind != Signature
func (a *Aster) Params() *types.Tuple {
	return a.signature().Params()
}

// Recv returns the receiver of signature s (if a method), or nil if a
// function. It is ignored when comparing signatures for identity.
//
// For an abstract method, Recv returns the enclosing interface either
// as a *Named or an *Interface. Due to embedding, an interface may
// contain methods whose receiver type is a different interface.
// NOTE: Panic, if TypKind != Signature
func (a *Aster) Recv() *types.Var {
	return a.signature().Recv()
}

// Results returns the results of signature s, or nil.
// NOTE: Panic, if TypKind != Signature
func (a *Aster) Results() *types.Tuple {
	return a.signature().Results()
}

// Variadic reports whether the signature s is variadic.
// NOTE: Panic, if TypKind != Signature
func (a *Aster) Variadic() bool {
	return a.signature().Variadic()
}
