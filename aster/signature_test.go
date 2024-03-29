// Copyright 2022 AndeyaLee. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aster_test

import (
	"go/ast"
	"testing"

	"github.com/andeya/aster/aster"
	"github.com/andeya/aster/tools"
	"github.com/stretchr/testify/assert"
)

func TestCoverBody1(t *testing.T) {
	var src = `package test
type M struct{}

const c = 1
func(m *M)M(){
	v:="M"
	_= v
}
var a=1
`
	const filename = "../_out/method.go"
	prog, _ := aster.LoadFile(filename, src)
	for _, s := range []string{
		`v:="new value1"
	_= v
	`,
		`_= "new\nvalue2"`,
		`a:=0
		a++
		a--`,
		`return "new error value3", errors.New("")`,
	} {
		prog.Inspect(func(fa aster.Facade) bool {
			if fa.ObjKind() != aster.Fun {
				return true
			}
			err := fa.CoverBody(s)
			if err != nil {
				t.Fatal(err)
			}
			body, err := fa.Body()
			if err != nil {
				t.Fatal(err)
			}
			t.Log(fa.Name(), "new_body:", body)
			return true
		})
		codes, err := prog.Format()
		if err != nil {
			t.Fatal(err)
		}
		t.Log(codes[filename])
	}
}

func TestCoverBody2(t *testing.T) {
	var src = `package test
type M struct{}

const c = 1
func(m *M)M()string{
	v:="M1"
	return v
}
var a=1
`
	const filename = "../_out/method.go"
	prog, _ := aster.LoadFile(filename, src)
	for _, s := range []string{
		`v:="new value1"
	return v
	`,
		`return "new\nvalue2"`,
		`a:=0
		a++
		a--
		return "new value3"
		`,
		`return "new error value4", nil`,
		`return "new error value5", errors.New("")`,
	} {
		prog.Inspect(func(fa aster.Facade) bool {
			if fa.ObjKind() != aster.Fun {
				return true
			}
			err := fa.CoverBody(s)
			if err != nil {
				t.Fatal(err)
			}
			body, err := fa.Body()
			if err != nil {
				t.Fatal(err)
			}
			t.Log(fa.Name(), "new_body:", body)
			return true
		})
		codes, err := prog.Format()
		if err != nil {
			t.Fatal(err)
		}
		t.Log(codes[filename])
	}
}

func TestMethod(t *testing.T) {
	var src = `package test
import "net/http/httputil"
type M struct{}

func(m *M)T(t httputil.BufferPool)(r *httputil.BufferPool){
	return *t
}
`
	const filename = "../_out/method.go"
	prog, _ := aster.LoadFile(filename, src)
	fa := prog.Lookup(aster.Typ, aster.Struct, "M")[0]
	method := fa.Method(0)
	t.Log(method.Body())
	assert.Equal(t, "T", method.Name())
	assert.Equal(t, "t", method.Params().At(0).Name())
	assert.Equal(t, "net/http/httputil.BufferPool", method.Params().At(0).Type().String())
	assert.Equal(t, "httputil.BufferPool", tools.CodeStyleType(method.Params().At(0).Type().String()))
	assert.Equal(t, "r", method.Results().At(0).Name())
	assert.Equal(t, "*net/http/httputil.BufferPool", method.Results().At(0).Type().String())
	assert.Equal(t, "*httputil.BufferPool", tools.CodeStyleType(method.Results().At(0).Type().String()))
	fnType := method.Node().(*ast.FuncDecl).Type
	reqType, err := method.FormatNode(fnType.Params.List[0].Type)
	assert.Nil(t, err)
	assert.Equal(t, "httputil.BufferPool", reqType)
	respType, err := method.FormatNode(fnType.Results.List[0].Type)
	assert.Nil(t, err)
	assert.Equal(t, "*httputil.BufferPool", respType)
}
