package aster_test

import (
	"fmt"
	"testing"

	"github.com/henrylee2cn/aster/aster"
)

func TestInspect(t *testing.T) {
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
	prog, err := aster.LoadFile("../_out/func1.go", src)
	if err != nil {
		t.Fatal(err)
	}
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
