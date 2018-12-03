# Aster [![GitHub release](https://img.shields.io/github/release/henrylee2cn/aster.svg?style=flat-square)](https://github.com/henrylee2cn/aster/releases) [![report card](https://goreportcard.com/badge/github.com/henrylee2cn/aster?style=flat-square)](http://goreportcard.com/report/henrylee2cn/aster) [![github issues](https://img.shields.io/github/issues/henrylee2cn/aster.svg?style=flat-square)](https://github.com/henrylee2cn/aster/issues?q=is%3Aopen+is%3Aissue) [![github closed issues](https://img.shields.io/github/issues-closed-raw/henrylee2cn/aster.svg?style=flat-square)](https://github.com/henrylee2cn/aster/issues?q=is%3Aissue+is%3Aclosed) [![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](http://godoc.org/github.com/henrylee2cn/aster) [![view examples](https://img.shields.io/badge/learn%20by-examples-00BCD4.svg?style=flat-square)](https://github.com/henrylee2cn/aster/tree/master/_examples)

Golang coding efficiency engine.

## Status

Under development, not for production...

## Feature

- Convert the AST to `reflect.Type`-like types(Kind-Flags), as it would at runtime
- Simpler, more natural way of metaprogramming
- Collect and package common syntax node types
- Provides easy-to-use traversal syntax node functionality
- Easily fetch and modify syntax node information
- ...

## An Example

- Set struct tag

```go
package main

import (
	"flag"
	"fmt"

	"github.com/henrylee2cn/aster/aster"
	"github.com/henrylee2cn/goutil"
)

var (
	filename = flag.String("filename", "out/eg.structtag.go", "file name")
	src      = flag.String("src", `package test
	type S struct {
		Apple string
		BananaPeel,car,OrangeWater string
		E int
	}
	func F(){
		type M struct {
			N int
			lowerCase string
		}
	}
	`, "code text")
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
	f, _ := aster.ParseFile(*filename, *src)
	f.Inspect(setStructTag)
	retCode, _ := f.Format()
    fmt.Println(retCode)
    _ = f.Store()
}
```

  - The output of the above program is:

	```go
	package test

	type S struct {
		Apple       string `json:"apple,omitempty"`
		BananaPeel  string `json:"banana_peel,omitempty"`
		car         string
		OrangeWater string `json:"orange_water,omitempty"`
		E           int    `json:"e,omitempty"`
	}

	func F() {
		type M struct {
			N         int `json:"n,omitempty"`
			lowerCase string
		}
	}
	```