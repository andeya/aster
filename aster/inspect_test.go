// Copyright 2018 henrylee2cn. All Rights Reserved.
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
	"testing"

	"github.com/henrylee2cn/aster/aster"
)

func TestLookup(t *testing.T) {
	var src = `package test
	type A int8
		const (
			F1 A = iota
			F2
		)
		func(a A)String()string {return string(a)}
		func B(int){
			b:=1
			_=b
		}
		var C=func()int{
			type c1 string
			var c2 = struct{}{}
			_=c2
			return 0
		}
		var D = struct{}{}
		type E = struct{}
	`
	prog, err := aster.LoadFile("../_out/lookup.go", src)
	if err != nil {
		t.Fatal(err)
	}
	pkg := prog.Package("test")
	list := pkg.Lookup(aster.Fun|aster.Var, aster.Signature|aster.Struct, "")
	for _, fa := range list {
		t.Log(fa)
	}
	list = pkg.Lookup(aster.Con, 0, "")
	for _, fa := range list {
		t.Log(fa)
	}
	list = pkg.Lookup(0, 0, "c1")
	for _, fa := range list {
		t.Log(fa)
	}
}
