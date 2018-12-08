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

package aster

import (
	"go/ast"
	"log"
)

func (p *PackageInfo) check() {
	log.Printf("Checking package %s...", p.String())
L:
	for ident, obj := range p.Defs {
		switch GetObjKind(obj) {
		case Bad, Lbl, Bui, Nil:
			continue
		case Var:
			nodes, _ := p.PathEnclosingInterval(ident.Pos(), ident.End())
			for _, n := range nodes {
				if _, ok := n.(*ast.Field); ok {
					continue L
				}
			}
		}
		p.addFacade(ident, obj)
	}
}

// Inspect traverses asters in the package.
func (p *PackageInfo) Inspect(fn func(*Facade) bool) {
	for _, fa := range p.facades {
		if !fn(fa) {
			return
		}
	}
}
