package main

import (
	"flag"
	"fmt"

	"github.com/henrylee2cn/aster/aster"
	"github.com/henrylee2cn/goutil"
)

var (
	filename = flag.String("filename", "out/eg.structtag.go", "file name")
	src      = flag.String("src", "package test", "code text")
)

func setStructTag(n aster.Node) bool {
	if n.Kind() != aster.Struct {
		return true
	}
	for i := n.NumField() - 1; i >= 0; i-- {
		field := n.Field(i)
		if !aster.IsExported(field.Name()) {
			continue
		}
		field.Tags.Set(&aster.Tag{
			Key:     "json",
			Name:    goutil.SnakeString(field.Name()),
			Options: []string{"omitempty"},
		})
	}
	return true
}

func main() {
	flag.Parse()

	f, err := aster.ParseFile(*filename, *src)
	if err != nil {
		panic(err)
	}

	f.Inspect(setStructTag)

	ret, err := f.Format()
	if err != nil {
		panic(err)
	}
	fmt.Println(ret)

	err = f.Store()
	if err != nil {
		panic(err)
	}
}
