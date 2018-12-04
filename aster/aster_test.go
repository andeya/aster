package aster_test

import (
	"go/format"
	"testing"

	"github.com/henrylee2cn/aster/aster"
)

func TestStruct(t *testing.T) {
	var src = []byte(`package test
// S comment
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
	t.Log(s)

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

func TestBasic(t *testing.T) {
	var src = []byte(`package test
	// A comment
	type A int
	// B comment
	type B *int
	// C comment
	type C []*int
	// D comment
	type D map[string]*int
	// E comment
	type E chan<- int
	// F comment
	type F = struct{}
	// G comment
	type G = *struct{}
`)
	src, err := format.Source(src)
	if err != nil {
		t.Fatal(err)
	}
	f, err := aster.ParseFile("../_out/alias1.go", src)
	if err != nil {
		t.Fatal(err)
	}
	f.Inspect(func(n aster.Node) bool {
		t.Log(n.Kind(), n)
		return true
	})
}

func TestFunc(t *testing.T) {
	var src = []byte(`package test
	// S comment
	type S struct {}
	// String comment
	func(s *S)String()string {return ""}
	// F1 comment
	func F1(i int){}
	// F2 FuncLit!
	var F2=func()int{}
`)
	src, err := format.Source(src)
	if err != nil {
		t.Fatal(err)
	}
	f, err := aster.ParseFile("../_out/func1.go", src)
	if err != nil {
		t.Fatal(err)
	}
	f.Inspect(func(n aster.Node) bool {
		if n.Kind() == aster.Func {
			t.Log(n.String())
		}
		return true
	})
}
