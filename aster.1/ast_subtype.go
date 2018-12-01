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
	"strconv"
)

// superType common type
type superType struct {
	kind    Kind
	pkgName func() string
	name    func() string
	methods []*Method
	doc     *ast.CommentGroup
	// numStars int
}

func newSuperType(pkgName, name func() string, kind Kind, doc *ast.CommentGroup) *superType {
	return &superType{
		kind:    kind,
		pkgName: pkgName,
		name:    name,
		doc:     doc,
	}
}

func (c *superType) addMethods(method ...*Method) {
	c.methods = append(c.methods, method...)
}

// Kind returns the facade kind of this type.
func (c *superType) Kind() Kind {
	return c.kind
}

// Name returns the type's name within its package for a defined type.
// For other (non-defined) types it returns the empty string.
func (c *superType) Name() string {
	return c.name()
}

// String returns a string representation of the type.
// The string representation may use shortened package names
// (e.g., base64 instead of "encoding/base64") and is not
// guaranteed to be unique among types. To test for type identity,
// compare the Types directly.
func (c *superType) String() string {
	if c.pkgName() == "" || c.name() == "" {
		return c.name()
	}
	return c.pkgName() + "." + c.name()
}

// Method returns the i'th method in the type's method set.
// It panics if i is not in the range [0, NumMethod()).
//
// For a non-interface type T or *T, the returned Method's Type and Func
// fields describe a function whose first argument is the receiver.
//
// For an interface type, the returned Method's Type field gives the
// method signature, without a receiver, and the Func field is nil.
func (c *superType) Method(i int) (*Method, bool) {
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
func (c *superType) MethodByName(name string) (*Method, bool) {
	for _, m := range c.methods {
		if m.Name == name {
			return m, true
		}
	}
	return nil, false
}

// NumMethod returns the number of exported methods in the type's method set.
func (c *superType) NumMethod() int {
	return len(c.methods)
}

// Implements reports whether the type implements the interface type u.
func (c *superType) Implements(u Type) bool {
	for i := u.NumMethod() - 1; i >= 0; i-- {
		um, _ := u.Method(i)
		cm, ok := c.MethodByName(um.Name)
		if !ok ||
			um.IsVariadic != cm.IsVariadic ||
			len(um.Params) != len(cm.Params) ||
			len(um.Results) != len(cm.Results) {
			return false
		}
		for k, v := range um.Params {
			if cm.Params[k].TypeName != v.TypeName {
				return false
			}
		}
		for k, v := range um.Results {
			if cm.Results[k].TypeName != v.TypeName {
				return false
			}
		}
	}
	return true
}

// Doc returns lead comment.
func (c *superType) Doc() string {
	if c.doc == nil {
		return ""
	}
	return c.doc.Text()
}

// SetDoc sets lead comment.
func (c *superType) SetDoc(text string) error {
	if c.Name() == "" {
		return errors.New("anonymous type cannot set document")
	}
	c.doc = &ast.CommentGroup{
		List: []*ast.Comment{{Text: text}},
	}
	return nil
}

// AliasType alias type such as `type T2 T` or `type T2 = T`
type AliasType struct {
	*superType
	ast.Node
	t        Type
	isAssign bool // is there `=` ?
	doc      *ast.CommentGroup
}

var _ Type = (*AliasType)(nil)

// NOTE: ast.Node is *ast.StarExpr, *ast.Ident or *ast.SelectorExpr
func newAliasType(node ast.Node, pkgName, name func() string, kind Kind, doc *ast.CommentGroup, isAssign bool) *AliasType {
	return &AliasType{
		superType: newSuperType(pkgName, name, kind, doc),
		Node:      node,
		isAssign:  isAssign,
		doc:       doc,
	}
}

// IsAssign is there `=` ?
func (a *AliasType) IsAssign() bool {
	return a.isAssign
}

// Origin returns the origin type.
func (a *AliasType) Origin() Type {
	return a.t
}

// ArrayType array type
type ArrayType struct {
	*superType
	*ast.ArrayType
}

func newArrayType(node *ast.ArrayType, pkgName, name func() string, doc *ast.CommentGroup) *ArrayType {
	return &ArrayType{
		superType: newSuperType(pkgName, name, Array, doc),
		ArrayType: node,
	}
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
	cnt, _ := strconv.Atoi(a.Node.(*ast.ArrayType).Len.(*ast.BasicLit).Value)
	return cnt
}

// SliceType slice
type SliceType struct {
	*superType
	elem Type
}

// TODO
func newSliceType(node *ast.ArrayType, name string, pkgName string, doc *ast.CommentGroup) *superType {
	return newSuperType(node, Slice, name, pkgName, doc)
}

// Kind returns the specific kind of this type.
func (*SliceType) Kind() Kind {
	return Array
}

// Elem returns a type's element type.
func (s *SliceType) Elem() Type {
	return s.elem
}

// MapType map
type MapType struct {
	*superType
	key   Type
	value Type
}

// TODO
func newMapType(node *ast.MapType, name string, pkgName string, doc *ast.CommentGroup) *superType {
	return newSuperType(node, Map, name, pkgName, doc)
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
	*superType
}

func newChanType(node *ast.ChanType, name string, pkgName string, doc *ast.CommentGroup) *superType {
	return newSuperType(node, Chan, name, pkgName, doc)
}

// ChanDir returns a channel type's direction.
// It panics if the type's Kind is not Chan.
func (c *ChanType) ChanDir() ast.ChanDir {
	return c.superType.Node.(*ast.ChanType).Dir
}

// InterfaceType represents a interface type.
type InterfaceType struct {
	*superType
}

func newInterfaceType(node *ast.InterfaceType, name string, pkgName string, doc *ast.CommentGroup, methods ...*Method) *superType {
	t := newSuperType(node, Interface, name, pkgName, doc)
	t.addMethods(methods...)
	return t
}

// FuncType function type
type FuncType struct {
	*superType
	params     []*FuncField
	results    []*FuncField
	isVariadic bool
}

func newFuncType(node *ast.FuncLit, name string, pkgName string, doc *ast.CommentGroup) *FuncType {
	f := &FuncType{
		superType:  newSuperType(node, Func, name, pkgName, doc),
		isVariadic: isVariadic(node.Type),
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
func (f *FuncType) Param(i int) (ff *FuncField, found bool) {
	if i < 0 || i >= len(f.params) {
		return
	}
	return f.params[i], true
}

// Result returns the type of a function type's i'th output parameter.
func (f *FuncType) Result(i int) (ff *FuncField, found bool) {
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
	*superType
	fields []*StructField // sorted by offset
}

func newStructType(node *ast.StructType, name string, pkgName string, doc *ast.CommentGroup) *StructType {
	return &StructType{
		superType: newSuperType(node, Struct, name, pkgName, doc),
	}
}

func (s *StructType) addFields(field ...*StructField) {
	s.fields = append(s.fields, field...)
}

// A StructField describes a single field in a struct.
type StructField struct {
	Name      string    // the field name
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
	BasicBool       Type = newBasicType("bool", Bool)
	BasicInt        Type = newBasicType("int", Int)
	BasicInt8       Type = newBasicType("int8", Int8)
	BasicInt16      Type = newBasicType("int16", Int16)
	BasicInt32      Type = newBasicType("int32", Int32)
	BasicInt64      Type = newBasicType("int64", Int64)
	BasicUint       Type = newBasicType("uint", Uint)
	BasicUint8      Type = newBasicType("uint8", Uint8)
	BasicUint16     Type = newBasicType("uint16", Uint16)
	BasicUint32     Type = newBasicType("uint32", Uint32)
	BasicUint64     Type = newBasicType("uint64", Uint64)
	BasicUintptr    Type = newBasicType("uintptr", Uintptr)
	BasicFloat32    Type = newBasicType("float32", Float32)
	BasicFloat64    Type = newBasicType("float64", Float64)
	BasicComplex64  Type = newBasicType("complex64", Complex64)
	BasicComplex128 Type = newBasicType("complex128", Complex128)
	BasicString     Type = newBasicType("string", String)
)

// BasicType basic type
type BasicType struct {
	ast.Node
	*superType
}

func newBasicType(name string, kind Kind) Type {
	return &BasicType{
		Node: NilNode,
		superType: newSuperType(
			func() string { return "" },
			func() string { return name },
			kind,
			nil,
		),
	}
}

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
	default:
		return nil, false
	}
	return
}
