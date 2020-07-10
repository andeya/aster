package aster_test

import (
	"github.com/henrylee2cn/aster/aster"
	"testing"
)

func TestFile_CoverImport(t *testing.T) {
	src := `// Package test for aster
package test
	import (
_ "aaa"
_ "errors"
_ "bbb"
)
`
	prog, _ := aster.LoadFile("../_out/inspect1.go", src)
	prog.PrintResume()
	pkg := prog.Package("test")
	pkg.Files[0].CoverImport("errors", "fmt", "_")
	err := pkg.Files[0].Rewrite()
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func TestFile_AddImport(t *testing.T) {
	src := `// Package test for aster
package test
	import (
_ "aaa"
_ "errors"
_ "bbb"
)
`
	prog, _ := aster.LoadFile("../_out/inspect2.go", src)
	prog.PrintResume()
	pkg := prog.Package("test")
	pkg.Files[0].AddImport("fmt", "_")
	err := pkg.Files[0].Rewrite()
	if err != nil {
		t.Fatalf(err.Error())
	}
}
