package aster_test

import (
	"fmt"
	"testing"

	"github.com/andeya/aster/aster"
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
	err1 := pkg.Files[0].AddImport("aaa", "_")
	fmt.Println(err1)
	err := pkg.Files[0].Rewrite()
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func TestFile_DelImport(t *testing.T) {
	src := `// Package test for aster
package test
	import (
_ "aaa"
_ "errors"
_ "bbb"
)
`
	prog, _ := aster.LoadFile("../_out/inspect3.go", src)
	prog.PrintResume()
	pkg := prog.Package("test")
	pkg.Files[0].DelImport("aaa")
	err := pkg.Files[0].Rewrite()
	if err != nil {
		t.Fatalf(err.Error())
	}
}
