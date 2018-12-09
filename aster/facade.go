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
	"go/types"
	"strings"
)

// An Facade describes a named language entity such as a package,
// constant, type, variable, function (incl. methods), or label.
// Facade interface implement all the objects.
//
// NOTE:
//  Objects of ObjKind=Bad are not collected;
//
type Facade interface {
	facadeIdentify() // only as identify

	// Ident returns the indent.
	Ident() *ast.Ident

	// Object returns the types.Object.
	Object() types.Object

	// ObjKind returns what the facade represents.
	ObjKind() ObjKind

	// TypKind returns what the facade type represents.
	// NOTE: If the type is *type.Named, returns the underlying TypKind.
	TypKind() TypKind

	// Id is a wrapper for Id(obj.Pkg(), obj.Name()).
	Id() string

	// Name returns the type's name within its package for a defined type.
	// For other (non-defined) types it returns the empty string.
	Name() string

	// Doc returns lead comment.
	Doc() string

	// CoverDoc covers lead comment if it exists.
	CoverDoc(text string) bool

	// Exported reports whether the object is exported (starts with a capital letter).
	// It doesn't take into account whether the object is in a local (function) scope
	// or not.
	Exported() bool

	// String previews the object formated code and comment.
	String() string

	// Underlying returns the underlying type of a type.
	Underlying() types.Type

	// IsAlias reports whether obj is an alias name for a type.
	IsAlias() bool

	// NumMethods returns the number of explicit methods whose receiver is named type t.
	NumMethods() int

	// Method returns the i'th method of named type t for 0 <= i < t.NumMethods().
	// NOTE: the result's TypKind is Signature.
	Method(i int) Facade

	// AssertableTo reports whether it can be asserted to have T's type.
	AssertableTo(T Facade) bool

	// AssignableTo reports whether it is assignable to a variable of T's type.
	AssignableTo(T Facade) bool

	// ConvertibleTo reports whether it is convertible to a value of T's type.
	ConvertibleTo(T Facade) bool

	// Implements reports whether it implements iface.
	// NOTE: Panic, if iface TypKind != Interface
	Implements(iface Facade, usePtr bool) bool

	// Elem returns the element type.
	// NOTE: Panic, if TypKind != (Array, Slice, Map, Chan and Pointer)
	Elem() types.Type

	// Key returns the key type of map.
	// NOTE: Panic, if TypKind != Map
	Key() types.Type

	// Len returns the length of array, or the number variables of tuple.
	// A negative result indicates an unknown length.
	// NOTE: Panic, if TypKind != Array and TypKind != Tuple
	Len() int64

	// ChanDir returns the direction of channel.
	// NOTE: Panic, if TypKind != Chan
	ChanDir() types.ChanDir

	// BasicInfo returns information about properties of basic type.
	// NOTE: Panic, if TypKind != Basic
	BasicInfo() types.BasicInfo

	// BasicKind returns the kind of basic type.
	// NOTE: Panic, if TypKind != Basic
	BasicKind() types.BasicKind

	// ----------------------------- TypKind = Signature (function) -----------------------------

	// IsMethod returns whether it is a method.
	IsMethod() bool

	// Params returns the parameters of signature s, or nil.
	// NOTE: Panic, if TypKind != Signature
	Params() *types.Tuple

	// Recv returns the receiver of signature s (if a method), or nil if a
	// function. It is ignored when comparing signatures for identity.
	//
	// For an abstract method, Recv returns the enclosing interface either
	// as a *Named or an *Interface. Due to embedding, an interface may
	// contain methods whose receiver type is a different interface.
	// NOTE: Panic, if TypKind != Signature
	Recv() *types.Var

	// Results returns the results of signature s, or nil.
	// NOTE: Panic, if TypKind != Signature
	Results() *types.Tuple

	// Variadic reports whether the signature s is variadic.
	// NOTE: Panic, if TypKind != Signature
	Variadic() bool

	// ---------------------------------- TypKind = Struct ----------------------------------

	// NumFields returns the number of fields in the struct (including blank and embedded fields).
	// NOTE: Panic, if TypKind != Struct
	NumFields() int

	// Field returns the i'th field for 0 <= i < NumFields().
	// NOTE:
	// Panic, if TypKind != Struct;
	// Panic, if i is not in the range [0, NumFields()).
	Field(i int) *StructField

	// FieldByName returns the struct field with the given name
	// and a boolean indicating if the field was found.
	// NOTE: Panic, if TypKind != Struct
	FieldByName(name string) (field *StructField, found bool)

	// ---------------------------------- TypKind = Interface ----------------------------------

	// EmbeddedType returns the i'th embedded type of interface fa for 0 <= i < fa.NumEmbeddeds().
	// NOTE: Panic, if TypKind != Interface
	IfaceEmbeddedType(i int) Facade

	// IfaceEmpty returns true if fa is the empty interface.
	IfaceEmpty() bool

	// IfaceExplicitMethod returns the i'th explicitly declared method of interface fa for 0 <= i < fa.NumExplicitMethods().
	// The methods are ordered by their unique Id.
	// NOTE:
	//  Panic, if TypKind != Interface;
	//  The result's TypKind is Signature.
	IfaceExplicitMethod(i int) Facade

	// IfaceNumEmbeddeds returns the number of embedded types in interface fa.
	// NOTE: Panic, if TypKind != Interface
	IfaceNumEmbeddeds() int

	// IfaceNumExplicitMethods returns the number of explicitly declared methods of interface fa.
	// NOTE: Panic, if TypKind != Interface
	IfaceNumExplicitMethods() int
}

type facade struct {
	obj          types.Object
	pkg          *PackageInfo
	ident        *ast.Ident
	doc          *ast.CommentGroup
	structFields []*StructField // effective only for structure
}

var _ Facade = (*facade)(nil)

func (fa *facade) facadeIdentify() {}

func (fa *facade) mustGetFacadeByObj(obj types.Object) *facade {
	facade, idx := fa.pkg.getFacadeByObj(obj)
	if idx < 0 {
		panic(fmt.Sprintf("aster: mustGetFacadeByObj can't find %s", obj.String()))
	}
	return facade
}

func (fa *facade) mustGetFacadeByTyp(typ types.Type) *facade {
	facade, idx := fa.pkg.getFacadeByTyp(typ)
	if idx < 0 {
		panic(fmt.Sprintf("aster: mustGetFacadeByTyp can't find %s", typ.String()))
	}
	return facade
}

// Ident returns the indent.
func (fa *facade) Ident() *ast.Ident {
	return fa.ident
}

// Object returns the types.Object.
func (fa *facade) Object() types.Object {
	return fa.obj
}

// ObjKind returns what the facade represents.
func (fa *facade) ObjKind() ObjKind {
	return GetObjKind(fa.obj)
}

// TypKind returns what the facade type represents.
// NOTE: If the type is *type.Named, returns the underlying TypKind.
func (fa *facade) TypKind() TypKind {
	return GetTypKind(fa.typ())
}

// typKind returns real TypKind.
func (fa *facade) typKind() TypKind {
	if fa.ObjKind() == Bad {
		return Invalid
	}
	return GetTypKind(fa.obj.Type())
}

func (fa *facade) typ() types.Type {
	if fa.typKind() == named {
		return fa.obj.Type().Underlying()
	}
	return fa.obj.Type()
}

// Id is a wrapper for Id(obj.Pkg(), obj.Name()).
func (fa *facade) Id() string { return fa.obj.Id() }

// Name returns the type's name within its package for a defined type.
// For other (non-defined) types it returns the empty string.
func (fa *facade) Name() string {
	return fa.ident.Name
}

// Doc returns lead comment.
func (fa *facade) Doc() string {
	return fa.doc.Text()
}

// CoverDoc covers lead comment if it exists.
func (fa *facade) CoverDoc(text string) bool {
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

// Exported reports whether the object is exported (starts with a capital letter).
// It doesn't take into account whether the object is in a local (function) scope
// or not.
func (fa *facade) Exported() bool { return fa.obj.Exported() }

// String previews the object formated code and comment.
func (fa *facade) String() string { return fa.pkg.Preview(fa.ident) }

// Underlying returns the underlying type of a type.
func (fa *facade) Underlying() types.Type {
	return fa.typ().Underlying()
}

// IsAlias reports whether obj is an alias name for a type.
func (fa *facade) IsAlias() bool {
	t, ok := fa.getNamed()
	if !ok {
		return false
	}
	return t.Obj().IsAlias()
}

func (fa *facade) getNamed() (*types.Named, bool) {
	if fa.typKind() != named {
		return nil, false
	}
	return fa.obj.Type().(*types.Named), true
}

// NumMethods returns the number of explicit methods whose receiver is named type t.
func (fa *facade) NumMethods() int {
	t, ok := fa.getNamed()
	if !ok {
		return 0
	}
	return t.NumMethods()
}

// Method returns the i'th method of named type t for 0 <= i < t.NumMethods().
// NOTE: the result's TypKind is Signature.
func (fa *facade) Method(i int) Facade {
	t, ok := fa.getNamed()
	if !ok {
		return nil
	}
	return fa.mustGetFacadeByObj(t.Method(i))
}

// AssertableTo reports whether it can be asserted to have T's type.
// NOTE: the current Facade's TypKind should be Interface.
func (fa *facade) AssertableTo(T Facade) bool {
	iface, ok := fa.typ().(*types.Interface)
	if !ok {
		return false
	}
	return types.AssertableTo(iface, T.(*facade).typ())
}

// AssignableTo reports whether it is assignable to a variable of T's type.
func (fa *facade) AssignableTo(T Facade) bool {
	return types.AssignableTo(fa.typ(), T.(*facade).typ())
}

// ConvertibleTo reports whether it is convertible to a value of T's type.
func (fa *facade) ConvertibleTo(T Facade) bool {
	return types.ConvertibleTo(fa.typ(), T.(*facade).typ())
}

// Implements reports whether it implements iface.
// NOTE: Panic, if iface TypKind != Interface
func (fa *facade) Implements(iface Facade, usePtr bool) bool {
	t := fa.obj.Type()
	if usePtr && fa.typKind() != Pointer {
		t = types.NewPointer(t)
	}
	return types.Implements(t, iface.(*facade).iface())
}
