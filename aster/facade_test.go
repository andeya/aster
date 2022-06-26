package aster_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/andeya/aster/aster"
)

var src = `// Package test for aster
package test
	import "errors"
	var err=errors.New("")
	// S comment
	type S int
	// String comment1
	// String comment2
	func(s S)String()string {return ""}
	func F1(i int){a:=func(){}}
	var F2=func()int{
		type G1 string
		var v = struct{}{}
		_=v
		return 0
	}
	var H1 = struct{}{}
	const X = 0
	type H2 = struct{X string}
	const (
		S1 S = iota
		S2
	)
	type H3 string
	type (
		H4 struct {Y string}
	)
	type H5 struct {
		Z string
	}
`

func TestFilename(t *testing.T) {
	want := "../_out/inspect1.go"
	prog, err := aster.LoadFile(want, src)
	if err != nil {
		t.Fatal(err)
	}
	prog.Inspect(func(fa aster.Facade) bool {
		if fa.File().Filename != want {
			t.Fatalf("want:%s, got:%s", want, fa.File().Filename)
		}
		return true
	})
	want, err = filepath.Abs("../_out/struct.go")
	if err != nil {
		t.Fatal(err)
	}

	prog, _ = aster.LoadDirs("../_out/")
	prog.Inspect(func(fa aster.Facade) bool {
		if fa.File().Filename != want {
			t.Fatalf("want:%s, got:%s", want, fa.File().Filename)
		}
		return true
	})

	prog, _ = aster.LoadPkgs("../_out/")
	prog.Inspect(func(fa aster.Facade) bool {
		if fa.File().Filename != want {
			t.Fatalf("want:%s, got:%s", want, fa.File().Filename)
		}
		return true
	})
}

func TestInspect(t *testing.T) {
	prog, _ := aster.LoadFile("../_out/inspect1.go", src)
	prog.PrintResume()
	pkg := prog.Package("test")
	var log string
	pkg.Inspect(func(fa aster.Facade) bool {
		log += fmt.Sprintf(
			"\nObjKind: %s\nTypKind: %s\nDoc: %sPreview:\n%s\nObj:\n%s\n",
			fa.ObjKind(), fa.TypKind(), fa.Doc(), fa, fa.Object(),
		)
		return true
	})
	t.Log(log)
}

func TestComment(t *testing.T) {
	prog, _ := aster.LoadFile("../_out/inspect1.go", src)
	prog.Inspect(func(fa aster.Facade) bool {
		succ := fa.SetDoc("aster: " + fa.Doc())
		if succ {
			t.Logf("Add doc comment prefix success: %s", fa.Id())
		} else {
			t.Logf("Add doc comment prefix fail: %s", fa.Id())
		}
		if fa.TypKind() == aster.Struct {
			for i := fa.NumFields() - 1; i >= 0; i-- {
				fa.Field(i).SetDoc("aster-field-doc...")
				fa.Field(i).SetComment("aster-field-comment...")
			}
		}
		return true
	})
	codes, _ := prog.Format()
	t.Log(codes["../_out/inspect1.go"])
}

// func TestAlias(t *testing.T) {
// 	var src = `package test
// 	// A comment
// 	type A int
// 	// B comment
// 	type B *int
// 	// C comment
// 	type C []*int
// 	// D comment
// 	type D map[string]*int
// 	// E comment
// 	type E chan<- int
// 	// F comment
// 	type F = struct{}
// 	// G comment
// 	type G = *struct{}
// `
// 	f, err := aster.ParseFile("../_out/alias1.go", src)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	f.Inspect(func(obj aster.Object) bool {
// 		t.Log(obj.ObjKind(), obj.Kind(), obj)
// 		return true
// 	})
// }

// func TestFunc(t *testing.T) {
// 	var src = `package test
// 	// S comment
// 	type S struct {}
// 	// String comment
// 	func(s *S)String()string {return ""}
// 	// F1 comment
// 	func F1(i int){}
// 	// F2 FuncLit!
// 	var F2=func()int{}
// `
// 	f, err := aster.ParseFile("../_out/func1.go", src)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	f.Inspect(func(n aster.Object) bool {
// 		if n.Kind() == aster.Func {
// 			t.Log(n.String())
// 		}
// 		return true
// 	})
// }
