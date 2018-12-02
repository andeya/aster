package test

import (
	"bytes"
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
	f, err := aster.ParseFile("", src)
	if err != nil {
		t.Fatal(err)
	}
	s, ok := f.LookupTypeBlock("S")
	if !ok {
		t.FailNow()
	}

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

	var dst bytes.Buffer
	err = format.Node(&dst, f.FileSet, s.Node())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(dst.String())

	ret, err := f.Format(f.File)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ret)
}
