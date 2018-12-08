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

func TestStruct(t *testing.T) {
	var src = `package test
type M []string
// S comment
type S1 struct {
	// a doc
	A string` + "`json:\"a\"`" + ` // a comment
	// bcd doc
	B,C,D int // line comment
	E int
	*M
}
var S2 = struct{
	F int
	// G comment
	G struct{
		H string
	}
	M
}{}
`
	prog, err := aster.LoadFile("../_out/struct.go", src)
	if err != nil {
		t.Fatal(err)
	}
	pkg := prog.Package("test")

	{
		s1 := pkg.Lookup(0, aster.Struct, "S1")[0]

		s1A, _ := s1.FieldByName("A")
		s1A.Tags().AddOptions("json", "omitempty")

		s1C, _ := s1.FieldByName("C")
		s1C.Tags().Set(&aster.Tag{
			Key:     "json",
			Name:    "c",
			Options: []string{"omitempty"},
		})

		s1M, _ := s1.FieldByName("M")
		s1M.Tags().Set(&aster.Tag{
			Key:  "json",
			Name: "m",
		})

		t.Log(s1)
	}
	{
		s2 := pkg.Lookup(0, aster.Struct, "S2")[0]

		s2G, _ := s2.FieldByName("G")
		s2G.Tags().Set(&aster.Tag{
			Key:     "json",
			Name:    "g",
			Options: []string{"omitempty"},
		})

		t.Log(s2)
	}
	{
		g := pkg.Lookup(0, aster.Struct, "G")[0]

		gH, _ := g.FieldByName("H")
		gH.Tags().Set(&aster.Tag{
			Key:  "json",
			Name: "h",
		})

		t.Log(g)
	}

	ret, err := pkg.Format()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ret)

	err = pkg.Rewrite()
	if err != nil {
		t.Fatal(err)
	}
}
