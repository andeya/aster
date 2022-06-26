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
	"go/types"
	"testing"

	"github.com/andeya/aster/aster"
)

func TestElem(t *testing.T) {
	var src = `package test
type A [3]int
var a = [3]int{1,2,3}

type B []string
var b = B{"a","b","c"}

type M map[string]bool
var m = M{"i":true, "j":false}

type C chan I
var c C = make(chan I, 3)

type P *S
var p P = new(S)

type U uint
var u uint = 3

type I interface{
	String()string
}
type S struct{}

func F(){}
`
	prog, _ := aster.LoadFile("../_out/mixer.go", src)
	prog.Inspect(func(fa aster.Facade) (next bool) {
		next = true
		var e types.Type
		switch kind := fa.TypKind(); kind {
		case aster.Array:
			e = fa.Elem()
			if e.String() != "int" {
				t.Fatalf("%v elem: want: %s, got: %v", fa, "int", e)
				return true
			}
		case aster.Slice:
			e = fa.Elem()
			if e.String() != "string" {
				t.Fatalf("%v elem: want: %s, got: %v", fa, "string", e)
				return true
			}
		case aster.Map:
			e = fa.Elem()
			if e.String() != "bool" {
				t.Fatalf("%v elem: want: %s, got: %v", fa, "bool", e)
				return true
			}
		case aster.Chan, aster.Pointer:
			e = fa.Elem()
			if kind == aster.Chan && e.String() != "test.I" {
				t.Fatalf("%v elem: want: %s, got: %v", fa, "test.I", e)
			} else if kind == aster.Pointer && e.String() != "test.S" {
				t.Fatalf("%v elem: want: %s, got: %v", fa, "test.S", e)
			} else if efa, found := prog.FindFacade(e); !found {
				t.Fatalf("FindFacade: not found %v", e)
			} else {
				t.Logf("%v elem: %v, facade: %v", fa, e, efa)
			}
			return true
		case aster.Basic, aster.Interface, aster.Struct, aster.Signature:
			defer func() {
				if p := recover(); p != nil {
					t.Logf("%v elem unsupport: %v", fa, p)
				} else {
					t.Fatalf("%v elem: want: unsupport, got: %v", fa, e)
				}
			}()
			e = fa.Elem()
			return true
		}
		t.Logf("%v elem: %v", fa, e)
		return true
	})
}
