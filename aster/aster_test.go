package aster_test

import (
	"testing"

	"github.com/henrylee2cn/aster/aster"
)

func TestStruct(t *testing.T) {
	var src = `package test
// S comment
type S struct {
	// a doc
	A string` + "`json:\"a\"`" + ` // a comment
	// bcd doc
	B,C,D int // line comment
	E int
}
`
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

func TestAlias(t *testing.T) {
	var src = `package test
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
`
	f, err := aster.ParseFile("../_out/alias1.go", src)
	if err != nil {
		t.Fatal(err)
	}
	f.Inspect(func(obj aster.Object) bool {
		t.Log(obj.ObjKind(), obj.Kind(), obj)
		return true
	})
}

func TestFunc(t *testing.T) {
	var src = `package test
	// S comment
	type S struct {}
	// String comment
	func(s *S)String()string {return ""}
	// F1 comment
	func F1(i int){}
	// F2 FuncLit!
	var F2=func()int{}
`
	f, err := aster.ParseFile("../_out/func1.go", src)
	if err != nil {
		t.Fatal(err)
	}
	f.Inspect(func(n aster.Object) bool {
		if n.Kind() == aster.Func {
			t.Log(n.String())
		}
		return true
	})
	pf, ok := f.LookupPureFunc("F2")
	if !ok {
		t.Fatal("not found F2")
	}
	t.Log(pf)
}

func TestScope(t *testing.T) {
	var src = `package test
	// S comment
	type S int
	// String comment
	func(s *S)String()string {return ""}
	// F1 comment
	func F1(i int){a:=func(){}}
	// F2 FuncLit!
	var F2=func()int{
		type G1 string
		var v = struct{}{}
		_=v
		return 0
	}
	// H1 comment
	var H1 = struct{}{}
	// H2 comment
	type H2 = struct{}
	const (
		S1 S = iota
		S2
	)
`
	f, err := aster.ParseFile("../_out/func1.go", src)
	if err != nil {
		t.Fatal(err)
	}

	f.Inspect(func(obj aster.Object) bool {
		t.Logf("ObjKind:%s, Kind:%s, IsGlobal:%v, Preview:%s", obj.ObjKind(), obj.Kind(), obj.IsGlobal(), obj)
		return true
	})
}
