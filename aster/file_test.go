package aster_test

import (
	"fmt"
	"testing"

	"github.com/andeya/aster/aster"
	"github.com/stretchr/testify/assert"
)

func TestFile_FindImport(t *testing.T) {
	src := `// Package test for aster
package test
import (
	_ "aaa"
	"bbb"
	_ "ccc"
	d "ddd"
)
`
	prog, _ := aster.LoadFile("../_out/inspect1.go", src)
	prog.PrintResume()
	f := prog.Package("test").Files[0]
	alias, found := f.FindImportByPath("bbb")
	assert.True(t, found)
	assert.Equal(t, "", alias)
	path, found := f.FindImportByAlias("bbb")
	assert.False(t, found)
	assert.Equal(t, "", path)
	path, found = f.FindImportByAlias("_")
	assert.True(t, found)
	assert.Equal(t, "aaa", path)
	path, found = f.FindImportByAlias("d")
	assert.True(t, found)
	assert.Equal(t, "ddd", path)
}

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
