# Aster [![Build Status](https://travis-ci.org/andeya/aster.svg?branch=master)](https://travis-ci.org/andeya/aster) <!-- [![Coverage Status](https://coveralls.io/repos/github/andeya/aster/badge.svg?branch=master)](https://coveralls.io/github/andeya/aster?branch=master) --> [![Report Card](https://goreportcard.com/badge/github.com/andeya/aster)](http://goreportcard.com/report/andeya/aster) [![GitHub release](https://img.shields.io/github/release/andeya/aster.svg)](https://github.com/andeya/aster/releases) [![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/andeya/aster)

轻松获取 golang 语法树并修改代码。

## 特性

- 将 AST 封装为类似 `reflect.Type` 的对象（Kind 标记），就像运行时反射一样
- 收集并封装常用类型的语法节点
- 提供易用的语法节点遍历功能
- 轻松修改语法节点信息
- 格式化输出修改后的代码
- 更简单、更自然的元编程方式
- ...

## Go 版本

- ≥go1.11

## 一个例子

- 设置 struct tag

```golang
package main

import (
	"flag"
	"fmt"

	"github.com/andeya/aster/aster"
	"github.com/andeya/goutil"
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

-  上面程序的输出：

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