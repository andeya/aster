package test

import (
	"go/format"
	"testing"

	"github.com/henrylee2cn/aster/aster"
)

func TestStruct(t *testing.T) {
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
	f, err := aster.ParseFile("", src)
	if err != nil {
		t.Fatal(err)
	}
	s, ok := f.LookupTypeInFile("S")
	if !ok {
		t.FailNow()
	}

	// test tag
	aField, ok := s.(*aster.StructType).FieldByName("A")
	if !ok {
		t.FailNow()
	}

	aField.Tags.AddOptions("json", "omitempty")

	bField, ok := s.(*aster.StructType).FieldByName("B")
	if !ok {
		t.FailNow()
	}
	bField.Tags.Set(&aster.Tag{
		Key:     "json",
		Name:    "bb",
		Options: []string{"omitempty"},
	})

	cField, _ := s.(*aster.StructType).FieldByName("C")

	ret, err := f.Format(f.File)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ret)

	// BUG: test comment
	aField.SetComment("a comment")
	t.Logf("S.A comment: %s", aField.Comment())
	t.Logf("S.A doc: %s", aField.Doc())
	t.Logf("S.C doc: %s", cField.Doc())
}
