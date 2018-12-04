package aster_test

import (
	"go/format"
	"testing"

	"github.com/henrylee2cn/aster/aster"
)

func TestStruct(t *testing.T) {
	var src = []byte(`package test
type S struct {
	// a doc
	A string` + "`json:\"a\"`" + ` // a comment
	// bcd doc
	B,C,D int // line comment
	E int
}
`)
	src, err := format.Source(src)
	if err != nil {
		t.Fatal(err)
	}
	f, err := aster.ParseFile("../_out/struct1.go", src)
	if err != nil {
		t.Fatal(err)
	}
	s, ok := f.LookupType("S")
	if !ok {
		t.FailNow()
	}
	t.Logf("package:%s, filename:%s, typename:%s", s.PkgName(), s.Filename(), s.Name())

	// test tag
	aField, ok := s.FieldByName("A")
	if !ok {
		t.FailNow()
	}
	aField.Tags.AddOptions("json", "omitempty")

	bField, ok := s.FieldByName("B")
	if !ok {
		t.FailNow()
	}
	bField.Tags.Set(&aster.Tag{
		Key:     "json",
		Name:    "b",
		Options: []string{"omitempty"},
	})

	dField, ok := s.FieldByName("D")
	if !ok {
		t.FailNow()
	}
	dField.Tags.Set(&aster.Tag{
		Key:     "json",
		Name:    "d",
		Options: []string{"omitempty"},
	})

	eField, ok := s.FieldByName("E")
	if !ok {
		t.FailNow()
	}
	eField.Tags.Set(&aster.Tag{
		Key:  "json",
		Name: "e",
	})

	structCode, err := f.FormatNode(s.Node())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(structCode)

	ret, err := f.Format()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ret)

	err = f.Store()
	if err != nil {
		t.Fatal(err)
	}
}
