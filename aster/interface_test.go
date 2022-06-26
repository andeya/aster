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
	"testing"

	"github.com/andeya/aster/aster"
)

func TestInterface(t *testing.T) {
	var src = `package test
type I1 interface{
	String()string
}
type I2 interface{
	String()string
	Print()
}
type M struct{}
func(m *M)String()string{return "M"}
`
	prog, _ := aster.LoadFile("../_out/interface.go", src)
	m := prog.Lookup(aster.Typ, aster.Struct, "M")[0]
	iface1 := prog.Lookup(aster.Typ, aster.Interface, "I1")[0]
	iface2 := prog.Lookup(aster.Typ, aster.Interface, "I2")[0]
	if !m.Implements(iface1, true) {
		t.Fatalf("type *M does not implement I1 interface")
	}
	if m.Implements(iface1, false) {
		t.Fatalf("type M implements I1 interface")
	}
	if m.Implements(iface2, true) {
		t.Fatalf("type *M implements I2 interface")
	}
	if m.Implements(iface2, false) {
		t.Fatalf("type M implements I2 interface")
	}
}
