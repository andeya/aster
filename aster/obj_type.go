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
	isGlobal bool
	isAssign bool // is there `=` for declared type?
	methods  []FuncObject
}

func (f *File) newSuperType(namePtr *string, kind Kind, isGlobal bool, doc *ast.CommentGroup,
	isAssign bool, objKind ...ast.ObjKind) *superType {
	a := ast.Typ
	if len(objKind) > 0 {
		a = objKind[0]
	}
	return &superType{
		super:    f.newSuper(namePtr, a, kind, doc),
		isGlobal: isGlobal,
		isAssign: isAssign,
	}
}

func (s *superType) typeObjectIdentify() {}

func (s *superType) IsGlobal() bool {
	return s.isGlobal
}

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
func (s *superType) Method(i int) (FuncObject, bool) {
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
func (s *superType) MethodByName(name string) (FuncObject, bool) {
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
func (s *superType) Implements(u TypeObject) bool {
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

func (s *superType) addMethod(method FuncObject) error {
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

// aliasType represents a alias type
type aliasType struct {
	*superType
	ast.Expr // type node
}

var _ Object = (*aliasType)(nil)
var _ TypeObject = (*aliasType)(nil)

func (f *File) newAliasType(namePtr *string, doc *ast.CommentGroup, assign token.Pos,
	typ ast.Expr) *basicType {
	kind := Suspense
	if _, ok := typ.(*ast.StarExpr); ok {
		kind = Ptr
	}
	return &basicType{
		superType: f.newSuperType(namePtr, kind, f.isGlobalTypOrFun(namePtr, typ), doc, assign != token.NoPos),
		Expr:      typ,
	}
}

// objType returns the node that declares the object type.
func (a *aliasType) objType() ast.Node {
	return a.Expr
}

// String returns the formated code block.
func (a *aliasType) String() string {
	return joinType(a, a.file)
}

// basicType represents a basic type
type basicType struct {
	*superType
	ast.Expr
}

var _ Object = (*basicType)(nil)
var _ TypeObject = (*basicType)(nil)

func (f *File) newBasicType(namePtr *string, doc *ast.CommentGroup, assign token.Pos,
	typ ast.Expr) (*basicType, bool) {
	basicName := strings.TrimLeft(f.TryFormatNode(typ), "*")
	kind, found := getBasicKind(basicName)
	if !found {
		return nil, false
	}
	return &basicType{
		superType: f.newSuperType(namePtr, kind, f.isGlobalTypOrFun(namePtr, typ), doc, assign != token.NoPos),
		Expr:      typ,
	}, true
}

func (f *File) newBasicOrAliasType(namePtr *string, doc *ast.CommentGroup, assign token.Pos,
	typ ast.Expr) Object {
	t, ok := f.newBasicType(namePtr, doc, assign, typ)
	if ok {
		return t
	}
	return f.newAliasType(namePtr, doc, assign, typ)
}

// objType returns the node that declares the object type.
func (b *basicType) objType() ast.Node {
	return b.Expr
}

// String returns the formated code block.
func (b *basicType) String() string {
	return joinType(b, b.file)
}

// listType represents an array or slice type.
type listType struct {
	*superType
	*ast.ArrayType
}

var _ Object = (*listType)(nil)
var _ TypeObject = (*listType)(nil)

func (f *File) newListType(namePtr *string, doc *ast.CommentGroup, assign token.Pos,
	typ *ast.ArrayType) *listType {
	kind := Slice
	if typ.Len != nil {
		kind = Array
	}
	return &listType{
		superType: f.newSuperType(namePtr, kind, f.isGlobalTypOrFun(namePtr, typ), doc, assign != token.NoPos),
		ArrayType: typ,
	}
}

// objType returns the node that declares the object type.
func (l *listType) objType() ast.Node {
	return l.ArrayType
}

// String returns the formated code block.
func (l *listType) String() string {
	return joinType(l, l.file)
}

// Len returns list's length if it is array type,
// otherwise returns false.
func (l *listType) Len() (int, bool) {
	if l.Kind() == Slice {
		return -1, false
	}
	cnt, _ := strconv.Atoi(l.ArrayType.Len.(*ast.BasicLit).Value)
	return cnt, true
}

// mapType represents a map type.
type mapType struct {
	*superType
	*ast.MapType
}

var _ Object = (*mapType)(nil)
var _ TypeObject = (*mapType)(nil)

func (f *File) newMapType(namePtr *string, doc *ast.CommentGroup, assign token.Pos,
	typ *ast.MapType) *mapType {
	return &mapType{
		superType: f.newSuperType(namePtr, Map, f.isGlobalTypOrFun(namePtr, typ), doc, assign != token.NoPos),
		MapType:   typ,
	}
}

// objType returns the node that declares the object type.
func (m *mapType) objType() ast.Node {
	return m.MapType
}

// String returns the formated code block.
func (m *mapType) String() string {
	return joinType(m, m.file)
}

// chanType represents a channel type.
type chanType struct {
	*superType
	*ast.ChanType
}

var _ Object = (*chanType)(nil)
var _ TypeObject = (*chanType)(nil)

func (f *File) newChanType(namePtr *string, doc *ast.CommentGroup, assign token.Pos,
	typ *ast.ChanType) *chanType {
	return &chanType{
		superType: f.newSuperType(namePtr, Chan, f.isGlobalTypOrFun(namePtr, typ), doc, assign != token.NoPos),
		ChanType:  typ,
	}
}

// objType returns the node that declares the object type.
func (c *chanType) objType() ast.Node {
	return c.ChanType
}

// String returns the formated code block.
func (c *chanType) String() string {
	return joinType(c, c.file)
}

// Dir returns a channel type's direction.
func (c *chanType) Dir() ast.ChanDir {
	return c.ChanType.Dir
}

// interfaceType represents a interface type.
type interfaceType struct {
	*superType
	*ast.InterfaceType
}

var _ Object = (*interfaceType)(nil)
var _ TypeObject = (*interfaceType)(nil)

func (f *File) newInterfaceType(namePtr *string, doc *ast.CommentGroup, assign token.Pos,
	typ *ast.InterfaceType) *interfaceType {
	return &interfaceType{
		superType:     f.newSuperType(namePtr, Interface, f.isGlobalTypOrFun(namePtr, typ), doc, assign != token.NoPos),
		InterfaceType: typ,
	}
}

// objType returns the node that declares the object type.
func (i *interfaceType) objType() ast.Node {
	return i.InterfaceType
}

// String returns the formated code block.
func (i *interfaceType) String() string {
	return joinType(i, i.file)
}

// structType represents a struct type.
type structType struct {
	*superType
	*ast.StructType
	fields []*StructField // sorted by offset
}

var _ Object = (*structType)(nil)
var _ TypeObject = (*structType)(nil)

func (f *File) newStructType(namePtr *string, objKind ast.ObjKind, doc *ast.CommentGroup, assign token.Pos,
	typ *ast.StructType) *structType {
	return &structType{
		superType: f.newSuperType(namePtr, Struct, f.isGlobalTypOrFun(namePtr, typ),
			doc, assign != token.NoPos, objKind),
		StructType: typ,
	}
}

// objType returns the node that declares the object type.
func (s *structType) objType() ast.Node {
	return s.StructType
}

// String returns the formated code block.
func (s *structType) String() string {
	return joinType(s, s.file)
}

// NumField returns a struct type's field count.
func (s *structType) NumField() int {
	return len(s.fields)
}

// Field returns a struct type's i'th field.
// It panics if the type's Kind is not Struct.
// It panics if i is not in the range [0, NumField()).
func (s *structType) Field(i int) (field *StructField) {
	if i < 0 || i >= len(s.fields) {
		panic("aster: Field index out of bounds")
	}
	return s.fields[i]
}

// FieldByName returns the struct field with the given name
// and a boolean indicating if the field was found.
func (s *structType) FieldByName(name string) (field *StructField, found bool) {
	for _, field := range s.fields {
		if field.Name() == name {
			return field, true
		}
	}
	return nil, false
}

// A StructField describes a single field in a struct.
type StructField struct {
	*ast.Field
	Tags *StructTag // field tags handler
}

func (s *structType) setFields() {
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

func joinType(obj Object, file *File) string {
	s, err := file.FormatNode(obj.objType())
	if err != nil {
		return fmt.Sprintf("// Formatting error: %s", err.Error())
	}
	doc := obj.Doc()
	if doc != "" {
		doc = "// " + doc
	}
	var assign string
	if obj.IsAssign() {
		assign = "= "
	}
	if obj.ObjKind() == ast.Var {
		return "var " + obj.Name() + " " + assign + s
	}
	return doc + "type " + obj.Name() + " " + assign + s
}
