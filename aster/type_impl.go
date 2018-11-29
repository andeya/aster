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
	"reflect"
)

// CommonType common type
type CommonType struct {
	ast.Node
	kind    Kind
	file    *File
	name    string
	methods []*Method
	declDoc *ast.CommentGroup
}

var _ Type = (*CommonType)(nil)

func newCommonType(kind Kind) *CommonType {
	switch kind {
	case Int,
		Int8,
		Int16,
		Int32,
		Int64,
		Uint,
		Uint8,
		Uint16,
		Uint32,
		Uint64,
		Uintptr,
		Float32,
		Float64,
		Complex64,
		Complex128,
		String,
		UnsafePointer,
		Interface:
		return &CommonType{kind: kind}
	}
	panic("CommonType must be a number, bool or string type")
}

// Kind returns the specific kind of this type.
func (c *CommonType) Kind() Kind {
	return c.kind
}

// Name returns the type's name within its package for a defined type.
// For other (non-defined) types it returns the empty string.
func (c *CommonType) Name() string {
	return c.name
}

// String returns a string representation of the type.
// The string representation may use shortened package names
// (e.g., base64 instead of "encoding/base64") and is not
// guaranteed to be unique among types. To test for type identity,
// compare the Types directly.
func (c *CommonType) String() string {
	if c.file.PkgName == "" {
		return c.name
	}
	return c.file.PkgName + "." + c.name
}

// Method returns the i'th method in the type's method set.
// It panics if i is not in the range [0, NumMethod()).
//
// For a non-interface type T or *T, the returned Method's Type and Func
// fields describe a function whose first argument is the receiver.
//
// For an interface type, the returned Method's Type field gives the
// method signature, without a receiver, and the Func field is nil.
func (c *CommonType) Method(i int) (*Method, bool) {
	if i < 0 || i >= len(c.methods) {
		return nil, false
	}
	return c.methods[i], true
}

// MethodByName returns the method with that name in the type's
// method set and a boolean indicating if the method was found.
//
// For a non-interface type T or *T, the returned Method's Type and Func
// fields describe a function whose first argument is the receiver.
//
// For an interface type, the returned Method's Type field gives the
// method signature, without a receiver, and the Func field is nil.
func (c *CommonType) MethodByName(name string) (*Method, bool) {
	for _, m := range c.methods {
		if m.Name == name {
			return m, true
		}
	}
	return nil, false
}

// NumMethod returns the number of exported methods in the type's method set.
func (c *CommonType) NumMethod() int {
	return len(c.methods)
}

// Implements reports whether the type implements the interface type u.
func (c *CommonType) Implements(u Type) bool {
	for i := u.NumMethod() - 1; i >= 0; i-- {
		um, _ := u.Method(i)
		cm, ok := c.MethodByName(um.Name)
		if !ok ||
			um.IsVariadic != cm.IsVariadic ||
			len(um.Params) != len(cm.Params) ||
			len(um.Result) != len(cm.Result) {
			return false
		}
		for k, v := range um.Params {
			if cm.Params[k].String() != v.String() {
				return false
			}
		}
		for k, v := range um.Result {
			if cm.Result[k].String() != v.String() {
				return false
			}
		}
	}
	return true
}

// Doc returns lead comment.
func (c *CommonType) Doc() string {
	if c.declDoc == nil {
		return ""
	}
	return c.declDoc.Text()
}

// SetDoc sets lead comment.
func (c *CommonType) SetDoc(text string) {
	c.declDoc = &ast.CommentGroup{
		List: []*ast.Comment{{Text: text}},
	}
}

// AliasType alias type such as `type T2 = T`
type AliasType struct {
	Type
	ast.Node
	file    *File
	name    string
	declDoc *ast.CommentGroup
}

// Name returns the type's name within its package for a defined type.
// For other (non-defined) types it returns the empty string.
func (a *AliasType) Name() string {
	return a.name
}

// String returns a string representation of the type.
// The string representation may use shortened package names
// (e.g., base64 instead of "encoding/base64") and is not
// guaranteed to be unique among types. To test for type identity,
// compare the Types directly.
func (a *AliasType) String() string {
	if a.file.PkgName == "" {
		return a.name
	}
	return a.file.PkgName + "." + a.name
}

// Doc returns lead comment.
func (a *AliasType) Doc() string {
	if a.declDoc == nil {
		return ""
	}
	return a.declDoc.Text()
}

// SetDoc sets lead comment.
func (a *AliasType) SetDoc(text string) {
	a.declDoc = &ast.CommentGroup{
		List: []*ast.Comment{{Text: text}},
	}
}

// Ref returns the reference type.
func (a *AliasType) Ref() Type {
	return a.Type
}

// CopyType type such as `type T2 T`
type CopyType struct {
	origin Type
	*CommonType
}

// Origin returns the origin type.
func (c *CopyType) Origin() Type {
	return c.origin
}

// PtrType pointer type
type PtrType struct {
	Type
}

func newPtrType(t Type) *PtrType {
	return &PtrType{t}
}

// Kind returns the specific kind of this type.
func (*PtrType) Kind() Kind {
	return Ptr
}

// String returns a string representation of the type.
// The string representation may use shortened package names
// (e.g., base64 instead of "encoding/base64") and is not
// guaranteed to be unique among types. To test for type identity,
// compare the Types directly.
func (p *PtrType) String() string {
	return "*" + p.Type.String()
}

// Elem returns a type's element type.
// It panics if the type's Kind is not Array, Chan, Map, Ptr, or Slice.
func (p *PtrType) Elem() Type {
	return p.Type
}

// SliceType slice
type SliceType struct {
	*CommonType
	elem Type
}

// Kind returns the specific kind of this type.
func (*SliceType) Kind() Kind {
	return Array
}

// Elem returns a type's element type.
func (s *SliceType) Elem() Type {
	return s.elem
}

// ArrayType array type
type ArrayType struct {
	*CommonType
	len  int
	elem Type
}

// Kind returns the specific kind of this type.
func (*ArrayType) Kind() Kind {
	return Array
}

// Elem returns a type's element type.
func (a *ArrayType) Elem() Type {
	return a.elem
}

// Len returns an array type's length.
func (a *ArrayType) Len() int {
	return a.len
}

// MapType map
type MapType struct {
	*CommonType
	key   Type
	value Type
}

// Kind returns the specific kind of this type.
func (*MapType) Kind() Kind {
	return Map
}

// Key returns the key type.
func (m *MapType) Key() Type {
	return m.key
}

// Value returns the key type.
func (m *MapType) Value() Type {
	return m.value
}

// ChanType represents a channel type's direction.
type ChanType struct {
	*CommonType
	chanDir reflect.ChanDir
}

// ChanDir returns a channel type's direction.
// It panics if the type's Kind is not Chan.
func (c *ChanType) ChanDir() reflect.ChanDir {
	return c.chanDir
}

// FuncType function type
type FuncType struct {
	*CommonType
	params     []Type
	result     []Type
	isVariadic bool
}

func newFuncType(node *ast.FuncDecl) *FuncType {
	f := &FuncType{}
	return f
}

// NumParam returns a function type's input parameter count.
func (f *FuncType) NumParam() int {
	return len(f.params)
}

// NumResult returns a function type's output parameter count.
func (f *FuncType) NumResult() int {
	return len(f.result)
}

// Param returns the type of a function type's i'th input parameter.
func (f *FuncType) Param(i int) (t Type, found bool) {
	if i < 0 || i >= len(f.params) {
		return
	}
	return f.params[i], true
}

// Result returns the type of a function type's i'th output parameter.
func (f *FuncType) Result(i int) (t Type, found bool) {
	if i < 0 || i >= len(f.result) {
		return
	}
	return f.result[i], true
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
	return f.isVariadic
}

// StructType represents a struct type.
type StructType struct {
	*CommonType
	fields []*StructField // sorted by offset
}

// A StructField describes a single field in a struct.
type StructField struct {
	Name      string    // the field name
	Exported  bool      // is upper case (exported) field name or not
	Type      Type      // field type
	Tag       StructTag // field tag string
	Index     []int     // index sequence for Type.FieldByIndex
	Anonymous bool      // is an embedded field
	Doc       string    // lead comment
	Comment   string    // line comment
}

// A StructTag is the tag string in a struct field.
//
// By convention, tag strings are a concatenation of
// optionally space-separated key:"value" pairs.
// Each key is a non-empty string consisting of non-control
// characters other than space (U+0020 ' '), quote (U+0022 '"'),
// and colon (U+003A ':').  Each value is quoted using U+0022 '"'
// characters and Go string literal syntax.
type StructTag = reflect.StructTag

// NumField returns a struct type's field count.
// It panics if the type's Kind is not Struct.
func (s *StructType) NumField() int {
	return len(s.fields)
}

// Field returns a struct type's i'th field.
func (s *StructType) Field(i int) (field *StructField, found bool) {
	if i < 0 || i >= len(s.fields) {
		return
	}
	return s.fields[i], true
}

// FieldByName returns the struct field with the given name
// and a boolean indicating if the field was found.
func (s *StructType) FieldByName(name string) (field *StructField, found bool) {
	for _, field := range s.fields {
		if field.Name == name {
			return field, true
		}
	}
	return nil, false
}
