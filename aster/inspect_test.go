package aster

import (
	"go/format"
	"testing"
)

func TestExpandFields(t *testing.T) {
	var src = []byte(`package test
type S struct {
	// xyz comment
	A,B,C int ` + "`json:\"xyz\"`" + `// abc comment
	D string
}
`)
	src, err := format.Source(src)
	if err != nil {
		t.Fatal(err)
	}
	f, err := ParseFile("", src)
	if err != nil {
		t.Fatal(err)
	}
	// s, ok := f.LookupTypeInFile("S")
	// if !ok {
	// 	t.FailNow()
	// }
	ret, err := f.Format(f.File)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ret)
}
