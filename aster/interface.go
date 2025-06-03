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

package aster

import (
	"fmt"
	"go/types"
)

// ---------------------------------- TypKind = Interface ----------------------------------

// NOTE: Panic, if TypKind != Interface
func (fa *facade) iface() *types.Interface {
	typ := fa.typ()
	t, ok := typ.(*types.Interface)
	if !ok {
		panic(fmt.Sprintf("aster: iface of non-Interface TypKind: %T", typ))
	}
	return t
}

// EmbeddedType returns the i'th embedded type of interface fa for 0 <= i < fa.NumEmbeddeds().
// NOTE: Panic, if TypKind != Interface
func (fa *facade) IfaceEmbeddedType(i int) Facade {
	t := fa.iface().EmbeddedType(i)
	return fa.mustGetFacadeByTyp(t)
}

// IfaceEmpty returns true if fa is the empty interface.
func (fa *facade) IfaceEmpty() bool {
	if iface, ok := fa.typ().(*types.Interface); ok {
		return iface.Empty()
	}
	return false
}

// IfaceExplicitMethod returns the i'th explicitly declared method of interface fa for 0 <= i < fa.NumExplicitMethods().
// The methods are ordered by their unique Id.
// NOTE:
//
//	Panic, if TypKind != Interface;
//	The result's TypKind is Signature.
func (fa *facade) IfaceExplicitMethod(i int) Facade {
	fn := fa.iface().ExplicitMethod(i)
	return fa.mustGetFacadeByObj(fn)
}

// IfaceNumEmbeddeds returns the number of embedded types in interface fa.
// NOTE: Panic, if TypKind != Interface
func (fa *facade) IfaceNumEmbeddeds() int {
	return fa.iface().NumEmbeddeds()
}

// IfaceNumExplicitMethods returns the number of explicitly declared methods of interface fa.
// NOTE: Panic, if TypKind != Interface
func (fa *facade) IfaceNumExplicitMethods() int {
	return fa.iface().NumExplicitMethods()
}
