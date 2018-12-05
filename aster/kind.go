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

//go:generate Stringer -type Kind

// A Kind represents the specific kind of an Object data.
// The zero Kind is not a valid kind.
type Kind uint

// Kind enumerate
const (
	Invalid Kind = iota
	Suspense
	Bool
	Int
	Int8
	Int16
	Int32
	Int64
	Uint
	Uint8
	Uint16
	Uint32
	Uint64
	Uintptr
	Float32
	Float64
	Complex64
	Complex128
	String
	Interface
	Chan
	Array
	Slice
	Map
	Func
	Struct
	Ptr
)

// IsPureFunc returns true if b is implementd FuncObject, but not method function.
func IsPureFunc(obj Object) bool {
	ok := obj.Kind() == Func
	if ok {
		_, ok = obj.Recv()
		return !ok
	}
	return false
}

// IsMethod returns true if b is implementd method FuncObject.
func IsMethod(obj Object) bool {
	ok := obj.Kind() == Func
	if ok {
		_, ok = obj.Recv()
	}
	return ok
}

// IsExported reports whether name is an exported Go symbol
// (that is, whether it begins with an upper-case letter).
//
func IsExported(name string) bool {
	return ast.IsExported(name)
}
