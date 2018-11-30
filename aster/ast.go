// Package aster is golang coding efficiency engine.
//
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
	"go/parser"
	"go/token"
	"os"
	"reflect"
)

// Module packages AST
type Module struct {
	FileSet *token.FileSet
	Dir     string
	filter  func(os.FileInfo) bool
	Pkgs    map[string]*Package // <package name, *Package>
	mode    parser.Mode
}

// A Package node represents a set of source files
// collectively building a Go package.
//
type Package struct {
	module  *Module // nil when not existed
	FileSet *token.FileSet
	Dir     string
	Name    string                 // package name
	Scope   *ast.Scope             // package scope across all files
	Imports map[string]*ast.Object // map of package id -> package object
	Files   map[string]*File       // Go source files by filename
	mode    parser.Mode
}

// A File node represents a Go source file.
//
// The Comments list contains all comments in the source file in order of
// appearance, including the comments that are pointed to from other nodes
// via Doc and Comment fields.
//
// For correct printing of source code containing comments (using packages
// go/format and go/printer), special care must be taken to update comments
// when a File's syntax tree is modified: For printing, comments are interspersed
// between tokens based on their position. If syntax tree nodes are
// removed or moved, relevant comments in their vicinity must also be removed
// (from the File.Comments list) or moved accordingly (by updating their
// positions). A CommentMap may be used to facilitate some of these operations.
//
// Whether and how a comment is associated with a node depends on the
// interpretation of the syntax tree by the manipulating program: Except for Doc
// and Comment comments directly associated with nodes, the remaining comments
// are "free-floating" (see also issues #18593, #20744).
//
type File struct {
	pkg      *Package // nil when not existed
	PkgName  string
	FileSet  *token.FileSet
	Filename string
	Src      []byte
	mode     parser.Mode
	Types    map[string]Type // <type name, Type>
	Imports  []*Import
	*ast.File
}

// Import import info
type Import struct {
	Name string
	Path string
	Doc  *ast.CommentGroup
}

// Type is the representation of a Go type.
type Type interface {
	ast.Node

	// Name returns the type's name within its package for a defined type.
	// For other (non-defined) types it returns the empty string.
	Name() string

	// String returns a string representation of the type.
	// The string representation may use shortened package names
	// (e.g., base64 instead of "encoding/base64") and is not
	// guaranteed to be unique among types. To test for type identity,
	// compare the Types directly.
	String() string

	// Kind returns the specific kind of this type.
	Kind() Kind

	// Method returns the i'th method in the type's method set.
	// For a non-interface type T or *T, the returned Method's Type and Func
	// fields describe a function whose first argument is the receiver.
	//
	// For an interface type, the returned Method's Type field gives the
	// method signature, without a receiver, and the Func field is nil.
	Method(int) (*Method, bool)

	// MethodByName returns the method with that name in the type's
	// method set and a boolean indicating if the method was found.
	//
	// For a non-interface type T or *T, the returned Method's Type and Func
	// fields describe a function whose first argument is the receiver.
	//
	// For an interface type, the returned Method's Type field gives the
	// method signature, without a receiver, and the Func field is nil.
	MethodByName(string) (*Method, bool)

	// NumMethod returns the number of exported methods in the type's method set.
	NumMethod() int

	// Implements reports whether the type implements the interface type u.
	Implements(u Type) bool

	// Doc returns lead comment.
	Doc() string

	// SetDoc sets lead comment.
	// NOTE: returns errror if Name==""
	SetDoc(string) error

	addMethods(method ...*Method)
}

// Method represents a single method.
type Method struct {
	*ast.FuncDecl
	Name       string // method name
	Recv       Type
	Params     []*FuncField
	Results    []*FuncField
	IsVariadic bool
	Doc        *ast.CommentGroup // lead comment
}

// FuncField function params or results.
type FuncField struct {
	Name     string
	TypeName string
}

// A Kind represents the specific kind of type that a Type represents.
// The zero Kind is not a valid kind.
type Kind = reflect.Kind

// Kind enumerate
const (
	Invalid Kind = reflect.Invalid

	// common types
	Bool       Kind = reflect.Bool
	Int        Kind = reflect.Int
	Int8       Kind = reflect.Int8
	Int16      Kind = reflect.Int16
	Int32      Kind = reflect.Int32
	Int64      Kind = reflect.Int64
	Uint       Kind = reflect.Uint
	Uint8      Kind = reflect.Uint8
	Uint16     Kind = reflect.Uint16
	Uint32     Kind = reflect.Uint32
	Uint64     Kind = reflect.Uint64
	Uintptr    Kind = reflect.Uintptr
	Float32    Kind = reflect.Float32
	Float64    Kind = reflect.Float64
	Complex64  Kind = reflect.Complex64
	Complex128 Kind = reflect.Complex128
	String     Kind = reflect.String

	// special types
	Interface Kind = reflect.Interface
	Chan      Kind = reflect.Chan
	Array     Kind = reflect.Array
	Slice     Kind = reflect.Slice
	Map       Kind = reflect.Map
	Func      Kind = reflect.Func
	Struct    Kind = reflect.Struct
	Ptr       Kind = reflect.Ptr
)

// NilNode nil Node
type NilNode struct{}

// Pos .
func (NilNode) Pos() token.Pos { return token.NoPos }

// End .
func (NilNode) End() token.Pos { return token.NoPos }
