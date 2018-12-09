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
	"sort"
	"strings"

	"github.com/henrylee2cn/structtag"
)

// ---------------------------------- TypKind = Struct ----------------------------------

// NOTE: Panic, if TypKind != Struct
func (fa *facade) structure() *types.Struct {
	typ := fa.typ()
	t, ok := typ.(*types.Struct)
	if !ok {
		panic(fmt.Sprintf("aster: structure of non-Struct TypKind: %T", typ))
	}
	// initiate
	if fa.structFields == nil {
		numFields := t.NumFields()
		fa.structFields = make([]*StructField, numFields)
		for expr, tv := range fa.pkg.info.Types {
			if tv.Type == t {
				n, ok := expr.(*ast.StructType)
				if !ok {
					n = expr.(*ast.CompositeLit).Type.(*ast.StructType)
				}
				expandFields(n.Fields)
				for i := 0; i < numFields; i++ {
					fa.structFields[i] = fa.pkg.newStructField(n.Fields.List[i], t.Field(i))
				}
				break
			}
		}
	}
	return t
}

// NumFields returns the number of fields in the struct (including blank and embedded fields).
// NOTE: Panic, if TypKind != Struct
func (fa *facade) NumFields() int {
	return fa.structure().NumFields()
}

// Field returns the i'th field for 0 <= i < NumFields().
// NOTE:
// Panic, if TypKind != Struct;
// Panic, if i is not in the range [0, NumFields()).
func (fa *facade) Field(i int) *StructField {
	fa.structure() // make sure initiated
	if i < 0 || i >= len(fa.structFields) {
		panic("aster: Field index out of bounds")
	}
	return fa.structFields[i]
}

// FieldByName returns the struct field with the given name
// and a boolean indicating if the field was found.
// NOTE: Panic, if TypKind != Struct
func (fa *facade) FieldByName(name string) (field *StructField, found bool) {
	fa.structure() // make sure initiated
	for _, field := range fa.structFields {
		if field.Name() == name {
			return field, true
		}
	}
	return nil, false
}

// StructField struct field object.
type StructField struct {
	node *ast.Field
	obj  *types.Var
	tags *Tags
}

func (p *PackageInfo) newStructField(node *ast.Field, obj *types.Var) *StructField {
	sf := &StructField{
		node: node,
		obj:  obj,
		tags: newTags(node),
	}
	return sf
}

// Name returns the field's name.
func (sf *StructField) Name() string {
	return sf.obj.Name()
}

// Exported reports whether the object is exported (starts with a capital letter).
// It doesn't take into account whether the object is in a local (function) scope
// or not.
func (sf *StructField) Exported() bool {
	return sf.obj.Exported()
}

// Tags returns the field's tag object.
func (sf *StructField) Tags() *Tags {
	return sf.tags
}

// Anonymous reports whether the variable is an embedded field.
// Same as Embedded; only present for backward-compatibility.
func (sf *StructField) Anonymous() bool {
	return sf.obj.Anonymous()
}

// Embedded reports whether the variable is an embedded field.
func (sf *StructField) Embedded() bool {
	return sf.obj.Embedded()
}

// Doc returns lead comment.
func (sf *StructField) Doc() string {
	if sf.node.Doc == nil {
		return ""
	}
	return sf.node.Doc.Text()
}

// Comment returns line comment.
func (sf *StructField) Comment() string {
	if sf.node.Comment == nil {
		return ""
	}
	return sf.node.Comment.Text()
}

// A Tags is the tag string in a struct field.
//
// By convention, tag strings are a concatenation of
// optionally space-separated key:"value" pairs.
// Each key is a non-empty string consisting of non-control
// characters other than space (U+0020 ' '), quote (U+0022 '"'),
// and colon (U+003A ':').  Each value is quoted using U+0022 '"'
// characters and Go string literal syntax.
type Tags struct {
	field *ast.Field
	tags  *structtag.Tags
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

func newTags(field *ast.Field) *Tags {
	tags := &Tags{
		field: field,
	}
	tags.reparse()
	return tags
}

func (s *Tags) reparse() (err error) {
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

func (s *Tags) resetValue() {
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

// Tags returns a slice of tags. The order is the original tag order unless it
// was changed.
func (s *Tags) Tags() []*Tag {
	return s.tags.Tags()
}

// AddOptions adds the given option for the given key. If the option already
// exists it doesn't add it again.
func (s *Tags) AddOptions(key string, options ...string) {
	s.tags.AddOptions(key, options...)
	s.resetValue()
}

// Delete deletes the tag for the given keys
func (s *Tags) Delete(keys ...string) {
	s.tags.Delete(keys...)
	s.resetValue()
}

// DeleteOptions deletes the given options for the given key
func (s *Tags) DeleteOptions(key string, options ...string) {
	s.tags.DeleteOptions(key, options...)
	s.resetValue()
}

// Get returns the tag associated with the given key. If the key is present
// in the tag the value (which may be empty) is returned. Otherwise the
// returned value will be the empty string. The ok return value reports whether
// the tag exists or not (which the return value is nil).
func (s *Tags) Get(key string) (*Tag, error) {
	return s.tags.Get(key)
}

// Keys returns a slice of tag keys. The order is the original tag order unless it
// was changed.
func (s *Tags) Keys() []string {
	return s.tags.Keys()
}

// Set sets the given tag. If the tag key already exists it'll override it
func (s *Tags) Set(tag *Tag) error {
	err := s.tags.Set(tag)
	if err == nil {
		s.resetValue()
	}
	return err
}

// String reassembles the tags into a valid literal tag field representation
func (s *Tags) String() string {
	return s.tags.String()
}

func expandFields(fieldList *ast.FieldList) {
	if fieldList == nil {
		return
	}
	var list = make([]*ast.Field, 0, fieldList.NumFields())
	for _, field := range fieldList.List {
		list = append(list, field)
		if len(field.Names) > 1 {
			for _, name := range field.Names[1:] {
				list = append(list, &ast.Field{
					Names: []*ast.Ident{cloneIdent(name)},
					Type:  field.Type,
					Tag:   cloneBasicLit(field.Tag),
				})
			}
			field.Names = field.Names[:1]
		}
	}
	fieldList.List = list
}
