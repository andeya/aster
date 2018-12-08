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

func TestMethod(t *testing.T) {
	var src = `package test
type M struct{}

func(m *M)String()string{return "M"}
`
	prog, _ := aster.LoadFile("../_out/method.go", src)
	m := prog.Lookup(aster.Typ, aster.Struct, "M")[0]
	num := m.NumMethods()
	for i := 0; i < num; i++ {
		method := m.Method(i)
		t.Logf("IsMethod:%v, Preview:%s", method.IsMethod(), method)
	}
}
