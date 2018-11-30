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
	"errors"
	"go/ast"
	"reflect"
)

// CommonType common type
type CommonType struct {
	ast.Node
	kind    Kind
	pkgName string
	name    string
	methods []*Method
	doc     *ast.CommentGroup
}

var _ Type = (*CommonType)(nil)

func newCommonType(node ast.Node, kind Kind, name string, pkgName string, doc *ast.CommentGroup) *CommonType {
	if node == nil {
		node = NilNode{}
	}
	return &CommonType{
		Node:    node,
		kind:    kind,
		pkgName: pkgName,
		name:    name,
		doc:     doc,
	}
}

func (c *CommonType) addMethods(method ...*Method) {
	c.methods = append(c.methods, method...)
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
	if c.pkgName == "" || c.name == "" {
		return c.name
	}
	return c.pkgName + "." + c.name
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
	if c.doc == nil {
		return ""
	}
	return c.doc.Text()
}

// SetDoc sets lead comment.
func (c *CommonType) SetDoc(text string) error {
	if c.Name() == "" {
		return errors.New("anonymous type cannot set document")
	}
	c.doc = &ast.CommentGroup{
		List: []*ast.Comment{{Text: text}},
	}
	return nil
}

// AliasType alias type such as `type T2 = T`
type AliasType struct {
	Type
	ast.Node
	file *File
	name string
	doc  *ast.CommentGroup
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
	if a.doc == nil {
		return ""
	}
	return a.doc.Text()
}

// SetDoc sets lead comment.
func (a *AliasType) SetDoc(text string) {
	a.doc = &ast.CommentGroup{
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
	return &PtrType{
		Type: t,
	}
}

// Kind returns the specific kind of this type.
func (p *PtrType) Kind() Kind {
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
	results    []Type
	isVariadic bool
}

func newFuncType(node ast.Node, name string, pkgName string, doc *ast.CommentGroup) *FuncType {
	var t *ast.FuncType
	switch x := node.(type) {
	case *ast.FuncLit:
		t = x.Type
	case *ast.FuncDecl:
		t = x.Type
	}
	f := &FuncType{
		CommonType: newCommonType(node, Func, name, pkgName, doc),
		isVariadic: isVariadic(t),
	}
	return f
}

// NumParam returns a function type's input parameter count.
func (f *FuncType) NumParam() int {
	return len(f.params)
}

// NumResult returns a function type's output parameter count.
func (f *FuncType) NumResult() int {
	return len(f.results)
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
	return f.isVariadic
}

// StructType represents a struct type.
type StructType struct {
	*CommonType
	fields []*StructField // sorted by offset
}

func newStructType(node ast.Node, name string, pkgName string, doc *ast.CommentGroup) *StructType {
	return &StructType{
		CommonType: newCommonType(node, Struct, name, pkgName, doc),
	}
}

func (s *StructType) addFields(field ...*StructField) {
	s.fields = append(s.fields, field...)
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

// basic types
var (
	BasicBool          Type = newCommonType(nil, Bool, "bool", "", nil)
	BasicInt           Type = newCommonType(nil, Int, "int", "", nil)
	BasicInt8          Type = newCommonType(nil, Int8, "int8", "", nil)
	BasicInt16         Type = newCommonType(nil, Int16, "int16", "", nil)
	BasicInt32         Type = newCommonType(nil, Int32, "int32", "", nil)
	BasicInt64         Type = newCommonType(nil, Int64, "int64", "", nil)
	BasicUint          Type = newCommonType(nil, Uint, "uint", "", nil)
	BasicUint8         Type = newCommonType(nil, Uint8, "uint8", "", nil)
	BasicUint16        Type = newCommonType(nil, Uint16, "uint16", "", nil)
	BasicUint32        Type = newCommonType(nil, Uint32, "uint32", "", nil)
	BasicUint64        Type = newCommonType(nil, Uint64, "uint64", "", nil)
	BasicUintptr       Type = newCommonType(nil, Uintptr, "uintptr", "", nil)
	BasicFloat32       Type = newCommonType(nil, Float32, "float32", "", nil)
	BasicFloat64       Type = newCommonType(nil, Float64, "float64", "", nil)
	BasicComplex64     Type = newCommonType(nil, Complex64, "complex64", "", nil)
	BasicComplex128    Type = newCommonType(nil, Complex128, "complex128", "", nil)
	BasicString        Type = newCommonType(nil, String, "string", "", nil)
	BasicUnsafePointer Type = newCommonType(nil, UnsafePointer, "unsafe.Pointer", "", nil)
)

func getBasicType(name string) (t Type, found bool) {
	found = true
	switch name {
	case "bool":
		t = BasicBool
	case "int":
		t = BasicInt
	case "int8":
		t = BasicInt8
	case "int16":
		t = BasicInt16
	case "int32":
		t = BasicInt32
	case "int64":
		t = BasicInt64
	case "uint":
		t = BasicUint
	case "uint8":
		t = BasicUint8
	case "uint16":
		t = BasicUint16
	case "uint32":
		t = BasicUint32
	case "uint64":
		t = BasicUint64
	case "uintptr":
		t = BasicUintptr
	case "float32":
		t = BasicFloat32
	case "float64":
		t = BasicFloat64
	case "complex64":
		t = BasicComplex64
	case "complex128":
		t = BasicComplex128
	case "string":
		t = BasicString
	case "unsafe.Pointer":
		t = BasicUnsafePointer
	default:
		return nil, false
	}
	return
}
