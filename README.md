# Aster [![Build Status](https://travis-ci.org/henrylee2cn/aster.svg?branch=master)](https://travis-ci.org/henrylee2cn/aster) <!-- [![Coverage Status](https://coveralls.io/repos/github/henrylee2cn/aster/badge.svg?branch=master)](https://coveralls.io/github/henrylee2cn/aster?branch=master) --> [![Report Card](https://goreportcard.com/badge/github.com/henrylee2cn/aster)](http://goreportcard.com/report/henrylee2cn/aster) [![GitHub release](https://img.shields.io/github/release/henrylee2cn/aster.svg)](https://github.com/henrylee2cn/aster/releases) [![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/henrylee2cn/aster)

Golang coding efficiency engine.

[中文版](./README_ZH.md)

## Feature

- Convert the AST to `reflect.Type`-like types(Kind-Flags), as it would at runtime
- Simpler, more natural way of metaprogramming
- Collect and package common syntax node types
- Provides easy-to-use traversal syntax node functionality
- Easily fetch and modify syntax node information
- ...

## Go Version

- go1.11

## An Example

- Set struct tag

```golang
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
	prog, _ := aster.LoadFile(*filename, *src)
	prog.Inspect(setStructTag)
	_ = prog.Rewrite()
}
```

- The output of the above program is:

```golang
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

## Task List

- [x] Basic
- [x] Array
- [x] Slice
- [x] Struct
- [x] Pointer
- [x] Tuple
- [x] Signature // non-builtin function or method
- [x] Interface
- [x] Map
- [x] Chan