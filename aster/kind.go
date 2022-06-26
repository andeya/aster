// Copyright 2022 AndeyaLee. All Rights Reserved.
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
	"go/types"
)

//go:generate Stringer -type ObjKind,TypKind -output kind_string.go

// ObjKind describes what an object statement represents.
// Extension based on ast.ObjKind: Buil and Nil
type ObjKind uint32

// The list of possible object statement kinds.
const (
	Bad ObjKind = 1 << iota // for error handling
	Pkg                     // package
	Con                     // constant
	Typ                     // type
	Var                     // variable
	Fun                     // function or method
	Lbl                     // label
	Bui                     // builtin
	Nil                     // nil
)

// TypKind describes what an object type represents.
type TypKind uint32

// The list of possible object type kinds.
const (
	Invalid TypKind = 1 << iota // type is invalid
	Basic
	Array
	Slice
	Struct
	Pointer
	Tuple
	Signature // non-builtin function or method
	Interface
	Map
	Chan
	named
)

// any kinds
const (
	AnyObjKind = ^ObjKind(0) // any object kind
	AnyTypKind = ^TypKind(0) // any type kind
)

// In judges whether k is fully contained in set.
func (k ObjKind) In(set ObjKind) bool {
	return k&set == k
}

// In judges whether k is fully contained in set.
func (k TypKind) In(set TypKind) bool {
	return k&set == k
}

// GetObjKind returns what the types.Object represents.
func GetObjKind(obj types.Object) ObjKind {
	switch obj.(type) {
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

// GetTypKind returns what the types.Type represents.
func GetTypKind(typ types.Type) TypKind {
	switch typ.(type) {
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
		return named
	}
	return Invalid
}
