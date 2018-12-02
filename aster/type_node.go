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
	"go/token"
	"sort"
	"strconv"
	"strings"

	"github.com/henrylee2cn/structtag"
)

type superType struct {
	*super
	isAssign bool // is there `=` for declared type?
	methods  []FuncBlock
}

func (f *File) newSuperType(namePtr *string, kind Kind, doc *ast.CommentGroup,
	isAssign bool) *superType {
	return &superType{
		super:    f.newSuper(namePtr, kind, doc),
		isAssign: isAssign,
	}
}

func (s *superType) typeBlockIdentify() {}

// IsAssign is there `=` for declared type?
func (s *superType) IsAssign() bool {
	return s.isAssign
}

// Method returns the i'th method in the type's method set.
// It panics if i is not in the range [0, NumMethod()).
//
// For a non-interface type T or *T, the returned Method's Type and Func
// fields describe a function whose first argument is the receiver.
//
// For an interface type, the returned Method's Type field gives the
// method signature, without a receiver, and the Func field is nil.
func (s *superType) Method(i int) (FuncBlock, bool) {
	if i < 0 || i >= len(s.methods) {
		return nil, false
	}
	return s.methods[i], true
}

// MethodByName returns the method with that name in the type's
// method set and a boolean indicating if the method was found.
//
// For a non-interface type T or *T, the returned Method's Type and Func
// fields describe a function whose first argument is the receiver.
//
// For an interface type, the returned Method's Type field gives the
// method signature, without a receiver, and the Func field is nil.
func (s *superType) MethodByName(name string) (FuncBlock, bool) {
	for _, m := range s.methods {
		if m.Name() == name {
			return m, true
		}
	}
	return nil, false
}

// NumMethod returns the number of exported methods in the type's method set.
func (s *superType) NumMethod() int {
	return len(s.methods)
}

// Implements reports whether the type implements the interface type u.
func (s *superType) Implements(u TypeBlock) bool {
	for i := u.NumMethod() - 1; i >= 0; i-- {
		um, _ := u.Method(i)
		cm, ok := s.MethodByName(um.Name())
		if !ok ||
			um.IsVariadic() != cm.IsVariadic() ||
			um.NumParam() != cm.NumParam() ||
			um.NumResult() != cm.NumResult() {
			return false
		}
		for j := um.NumParam(); j >= 0; j-- {
			uf, _ := um.Param(j)
			cf, _ := cm.Param(j)
			if uf.TypeName != cf.TypeName {
				return false
			}
		}
		for j := um.NumResult(); j >= 0; j-- {
			uf, _ := um.Result(j)
			cf, _ := cm.Result(j)
			if uf.TypeName != cf.TypeName {
				return false
			}
		}
	}
	return true
}

func (s *superType) addMethod(method FuncBlock) error {
	field, ok := method.Recv()
	if !ok {
		return fmt.Errorf("not method: %s", method.Name())
	}
	if field.TypeName != s.Name() {
		return fmt.Errorf("reveiver do not match method: %s, want: %s, got: %s",
			method.Name(), s.Name(), field.TypeName)
	}
	s.methods = append(s.methods, method)
	return nil
}

// AliasType represents a alias type
type AliasType struct {
	*superType
	ast.Expr // type node
}

var _ Block = (*AliasType)(nil)
var _ TypeBlock = (*AliasType)(nil)

func (f *File) newAliasType(namePtr *string, doc *ast.CommentGroup, assign token.Pos,
	typ ast.Expr) *BasicType {
	return &BasicType{
		superType: f.newSuperType(namePtr, Suspense, doc, assign != token.NoPos),
		Expr:      typ,
	}
}

// Node returns origin AST node.
func (a *AliasType) Node() ast.Node {
	return a.Expr
}

// BasicType represents a basic type
type BasicType struct {
	*superType
	ast.Expr
}

var _ Block = (*BasicType)(nil)
var _ TypeBlock = (*BasicType)(nil)

func (f *File) newBasicType(namePtr *string, doc *ast.CommentGroup, assign token.Pos,
	typ ast.Expr) (*BasicType, bool) {
	basicName := strings.TrimLeft(f.TryFormat(typ), "*")
	kind, found := getBasicKind(basicName)
	if !found {
		return nil, false
	}
	return &BasicType{
		superType: f.newSuperType(namePtr, kind, doc, assign != token.NoPos),
		Expr:      typ,
	}, true
}

func (f *File) newBasicOrAliasType(namePtr *string, doc *ast.CommentGroup, assign token.Pos,
	typ ast.Expr) Block {
	t, ok := f.newBasicType(namePtr, doc, assign, typ)
	if ok {
		return t
	}
	return f.newAliasType(namePtr, doc, assign, typ)
}

// Node returns origin AST node.
func (b *BasicType) Node() ast.Node {
	return b.Expr
}

// ListType represents an array or slice type.
type ListType struct {
	*superType
	*ast.ArrayType
}

var _ Block = (*ListType)(nil)
var _ TypeBlock = (*ListType)(nil)

func (f *File) newListType(namePtr *string, doc *ast.CommentGroup, assign token.Pos,
	typ *ast.ArrayType) *ListType {
	kind := Slice
	if typ.Len != nil {
		kind = Array
	}
	return &ListType{
		superType: f.newSuperType(namePtr, kind, doc, assign != token.NoPos),
		ArrayType: typ,
	}
}

// Node returns origin AST node.
func (l *ListType) Node() ast.Node {
	return l.ArrayType
}

// Len returns list's length if it is array type,
// otherwise returns false.
func (l *ListType) Len() (int, bool) {
	if l.Kind() == Slice {
		return -1, false
	}
	cnt, _ := strconv.Atoi(l.ArrayType.Len.(*ast.BasicLit).Value)
	return cnt, true
}

// MapType represents a map type.
type MapType struct {
	*superType
	*ast.MapType
}

var _ Block = (*MapType)(nil)
var _ TypeBlock = (*MapType)(nil)

func (f *File) newMapType(namePtr *string, doc *ast.CommentGroup, assign token.Pos,
	typ *ast.MapType) *MapType {
	return &MapType{
		superType: f.newSuperType(namePtr, Map, doc, assign != token.NoPos),
		MapType:   typ,
	}
}

// Node returns origin AST node.
func (m *MapType) Node() ast.Node {
	return m.MapType
}

// ChanType represents a channel type.
type ChanType struct {
	*superType
	*ast.ChanType
}

var _ Block = (*ChanType)(nil)
var _ TypeBlock = (*ChanType)(nil)

func (f *File) newChanType(namePtr *string, doc *ast.CommentGroup, assign token.Pos,
	typ *ast.ChanType) *ChanType {
	return &ChanType{
		superType: f.newSuperType(namePtr, Chan, doc, assign != token.NoPos),
		ChanType:  typ,
	}
}

// Node returns origin AST node.
func (c *ChanType) Node() ast.Node {
	return c.ChanType
}

// Dir returns a channel type's direction.
func (c *ChanType) Dir() ast.ChanDir {
	return c.ChanType.Dir
}

// InterfaceType represents a interface type.
type InterfaceType struct {
	*superType
	*ast.InterfaceType
}

var _ Block = (*InterfaceType)(nil)
var _ TypeBlock = (*InterfaceType)(nil)

func (f *File) newInterfaceType(namePtr *string, doc *ast.CommentGroup, assign token.Pos,
	typ *ast.InterfaceType) *InterfaceType {
	return &InterfaceType{
		superType:     f.newSuperType(namePtr, Interface, doc, assign != token.NoPos),
		InterfaceType: typ,
	}
}

// Node returns origin AST node.
func (i *InterfaceType) Node() ast.Node {
	return i.InterfaceType
}

// StructType represents a struct type.
type StructType struct {
	*superType
	*ast.StructType
	fields []*StructField // sorted by offset
}

var _ Block = (*StructType)(nil)
var _ TypeBlock = (*StructType)(nil)

func (f *File) newStructType(namePtr *string, doc *ast.CommentGroup, assign token.Pos,
	typ *ast.StructType) *StructType {
	return &StructType{
		superType:  f.newSuperType(namePtr, Struct, doc, assign != token.NoPos),
		StructType: typ,
	}
}

// Node returns origin AST node.
func (s *StructType) Node() ast.Node {
	return s.StructType
}

// A StructField describes a single field in a struct.
type StructField struct {
	*ast.Field
	Tags *StructTag // field tags handler
}

func (s *StructType) setFields() {
	expandFields(s.StructType.Fields)
	for _, field := range s.StructType.Fields.List {
		s.fields = append(s.fields, &StructField{
			Field: field,
			Tags:  newStructTag(field),
		})
	}
}

// Name returns field name
func (s *StructField) Name() string {
	if !s.Anonymous() {
		return s.Field.Names[0].Name
	}
	ident, _ := getElem(s.Field.Type).(*ast.Ident)
	if ident == nil {
		return ""
	}
	return ident.Name
}

// Doc returns lead comment.
func (s *StructField) Doc() string {
	if s.Field.Doc == nil {
		return ""
	}
	return s.Field.Doc.Text()
}

// Comment returns line comment.
func (s *StructField) Comment() string {
	if s.Field.Comment == nil {
		return ""
	}
	return s.Field.Comment.Text()
}

// Anonymous returns whether the field is an anonymous field.
func (s *StructField) Anonymous() bool {
	return len(s.Field.Names) == 0
}

// NumField returns a struct type's field count.
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
		if field.Name() == name {
			return field, true
		}
	}
	return nil, false
}

// A StructTag is the tag string in a struct field.
//
// By convention, tag strings are a concatenation of
// optionally space-separated key:"value" pairs.
// Each key is a non-empty string consisting of non-control
// characters other than space (U+0020 ' '), quote (U+0022 '"'),
// and colon (U+003A ':').  Each value is quoted using U+0022 '"'
// characters and Go string literal syntax.
type StructTag struct {
	field *ast.Field
	tags  *structtag.Tags
}

func newStructTag(field *ast.Field) *StructTag {
	tags := &StructTag{
		field: field,
	}
	tags.reparse()
	return tags
}

func (s *StructTag) reparse() (err error) {
	var value string
	if s.field.Tag != nil {
		value = strings.Trim(s.field.Tag.Value, "`")
	}
	s.tags, err = structtag.Parse(value)
	if err != nil {
		s.tags, _ = structtag.Parse("")
	}
	return err
}

func (s *StructTag) resetValue() {
	sort.Sort(s.tags)
	value := s.tags.String()
	if value == "" {
		s.field.Tag = nil
	} else {
		if s.field.Tag == nil {
			s.field.Tag = &ast.BasicLit{}
		}
		s.field.Tag.Value = "`" + value + "`"
	}
}

// Tag defines a single struct's string literal tag
//
// type Tag struct {
// Key is the tag key, such as json, xml, etc..
// i.e: `json:"foo,omitempty". Here key is: "json"
// Key string
//
// Name is a part of the value
// i.e: `json:"foo,omitempty". Here name is: "foo"
// Name string
//
// Options is a part of the value. It contains a slice of tag options i.e:
// `json:"foo,omitempty". Here options is: ["omitempty"]
// Options []string
// }
//
type Tag = structtag.Tag

// Tags returns a slice of tags. The order is the original tag order unless it
// was changed.
func (s *StructTag) Tags() []*Tag {
	return s.tags.Tags()
}

// AddOptions adds the given option for the given key. If the option already
// exists it doesn't add it again.
func (s *StructTag) AddOptions(key string, options ...string) {
	s.tags.AddOptions(key, options...)
	s.resetValue()
}

// Delete deletes the tag for the given keys
func (s *StructTag) Delete(keys ...string) {
	s.tags.Delete(keys...)
	s.resetValue()
}

// DeleteOptions deletes the given options for the given key
func (s *StructTag) DeleteOptions(key string, options ...string) {
	s.tags.DeleteOptions(key, options...)
	s.resetValue()
}

// Get returns the tag associated with the given key. If the key is present
// in the tag the value (which may be empty) is returned. Otherwise the
// returned value will be the empty string. The ok return value reports whether
// the tag exists or not (which the return value is nil).
func (s *StructTag) Get(key string) (*Tag, error) {
	return s.tags.Get(key)
}

// Keys returns a slice of tag keys. The order is the original tag order unless it
// was changed.
func (s *StructTag) Keys() []string {
	return s.tags.Keys()
}

// Set sets the given tag. If the tag key already exists it'll override it
func (s *StructTag) Set(tag *Tag) error {
	err := s.tags.Set(tag)
	if err == nil {
		s.resetValue()
	}
	return err
}

// String reassembles the tags into a valid literal tag field representation
func (s *StructTag) String() string {
	return s.tags.String()
}
