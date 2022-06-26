package main

import (
	"flag"
	"fmt"

	"github.com/andeya/aster/aster"
	"github.com/andeya/goutil"
)

var (
	filename = flag.String("filename", "out/eg.structtag.go", "file name")
	src      = flag.String("src", "package test", "code text")
)

func setStructTag(fa aster.Facade) bool {
	if fa.TypKind() != aster.Struct {
		return true
	}
	for i := fa.NumFields() - 1; i >= 0; i-- {
		field := fa.Field(i)
		if !field.Exported() {
			continue
		}
		field.Tags().Set(&aster.Tag{
			Key:     "json",
			Name:    goutil.SnakeString(field.Name()),
			Options: []string{"omitempty"},
		})
	}
	return true
}

func main() {
	flag.Parse()

	prog, err := aster.LoadFile(*filename, *src)
	if err != nil {
		panic(err)
	}

	prog.Inspect(setStructTag)

	ret, err := prog.Format()
	if err != nil {
		panic(err)
	}
	fmt.Println(ret)

	// err = prog.Rewrite()
	// if err != nil {
	// 	panic(err)
	// }
}
