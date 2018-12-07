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

//go:generate Stringer -type ObjKind,TypKind -output kind_string.go

// ObjKind describes what an object statement represents.
// Extension based on ast.ObjKind: Buil and Nil
type ObjKind int

// The list of possible object statement kinds.
const (
	Bad ObjKind = iota // for error handling
	Pkg                // package
	Con                // constant
	Typ                // type
	Var                // variable
	Fun                // function or method
	Lbl                // label
	Bui                // builtin
	Nil                // nil
)

// TypKind describes what an object type represents.
type TypKind int

// The list of possible object type kinds.
const (
	Invalid TypKind = iota // type is invalid
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
	Named
)
