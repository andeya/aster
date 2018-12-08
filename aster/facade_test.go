package aster_test

import (
	"fmt"
	"sort"
	"strings"
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
	printProgram(prog)
	pkg := prog.Package("test")
	var log string
	pkg.Inspect(func(fa *aster.Facade) bool {
		log += fmt.Sprintf(
			"\nObjKind: %s\nTypKind: %s\nDoc: %sPreview:\n%s\nObj:\n%s\n",
			fa.ObjKind(), fa.TypKind(), fa.Doc(), fa, fa.Object(),
		)
		return true
	})
	t.Log(log)
}

// func TestStruct(t *testing.T) {
// 	var src = `package test
// // S comment
// type S struct {
// 	// a doc
// 	A string` + "`json:\"a\"`" + ` // a comment
// 	// bcd doc
// 	B,C,D int // line comment
// 	E int
// }
// `
// 	f, err := aster.ParseFile("../_out/struct1.go", src)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	s, ok := f.LookupType("S")
// 	if !ok {
// 		t.FailNow()
// 	}
// 	t.Logf("package:%s, filename:%s, typename:%s", s.PkgName(), s.Filename(), s.Name())
// 	t.Log(s)

// 	// test tag
// 	aField, ok := s.FieldByName("A")
// 	if !ok {
// 		t.FailNow()
// 	}
// 	aField.Tags.AddOptions("json", "omitempty")

// 	bField, ok := s.FieldByName("B")
// 	if !ok {
// 		t.FailNow()
// 	}
// 	bField.Tags.Set(&aster.Tag{
// 		Key:     "json",
// 		Name:    "b",
// 		Options: []string{"omitempty"},
// 	})

// 	dField, ok := s.FieldByName("D")
// 	if !ok {
// 		t.FailNow()
// 	}
// 	dField.Tags.Set(&aster.Tag{
// 		Key:     "json",
// 		Name:    "d",
// 		Options: []string{"omitempty"},
// 	})

// 	eField, ok := s.FieldByName("E")
// 	if !ok {
// 		t.FailNow()
// 	}
// 	eField.Tags.Set(&aster.Tag{
// 		Key:  "json",
// 		Name: "e",
// 	})

// 	ret, err := f.Format()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	t.Log(ret)

// 	err = f.Store()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// }

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

func printProgram(prog *aster.Program) {
	// Created packages are the initial packages specified by a call
	// to CreateFromFilenames or CreateFromFiles.
	var names []string
	for _, info := range prog.Created {
		names = append(names, info.Pkg.Path())
	}
	fmt.Printf("created: %s\n", names)

	// Imported packages are the initial packages specified by a
	// call to Import or ImportWithTests.
	names = nil
	for _, info := range prog.Imported {
		if strings.Contains(info.Pkg.Path(), "internal") {
			continue // skip, to reduce fragility
		}
		names = append(names, info.Pkg.Path())
	}
	sort.Strings(names)
	fmt.Printf("imported: %s\n", names)

	// InitialPackages contains the union of created and imported.
	names = nil
	for _, info := range prog.InitialPackages() {
		names = append(names, info.Pkg.Path())
	}
	sort.Strings(names)
	fmt.Printf("initial: %s\n", names)

	// AllPackages contains all initial packages and their dependencies.
	names = nil
	for pkg := range prog.AllPackages {
		names = append(names, pkg.Path())
	}
	sort.Strings(names)
	fmt.Printf("all: %s\n", names)
}
