# Aster [![GitHub release](https://img.shields.io/github/release/henrylee2cn/aster.svg?style=flat-square)](https://github.com/henrylee2cn/aster/releases) [![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](http://godoc.org/github.com/henrylee2cn/aster) [![Report Card](https://goreportcard.com/badge/github.com/henrylee2cn/aster?style=flat-square)](http://goreportcard.com/report/henrylee2cn/aster) [![Build Status](https://travis-ci.org/henrylee2cn/aster.svg?branch=master)](https://travis-ci.org/henrylee2cn/aster)

Golang 高效编码引擎。

## 状态

正在开发中，不能用于生产...

## 特性

- 将 AST 封装为类似 `reflect.Type` 的对象（Kind 标记），就像运行时反射一样
- 更简单、更自然的元编程方式
- 收集并封装常用类型的语法节点
- 提供易用的语法节点遍历功能
- 轻松获取和修改语法节点信息
- ...

## 测试环境

- OS:
	+ Linux
	+ OSX
	+ Windows

- Go:
	+ 1.11

## 一个例子

- 设置 struct tag

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
