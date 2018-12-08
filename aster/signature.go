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

import "go/types"

// ---------------------------------- TypKind = Signature (function) ----------------------------------

// NOTE: Panic, if TypKind != Signature
func (fa *facade) signature() *types.Signature {
	return fa.typ().(*types.Signature)
}

// IsMethod returns whether it is a method.
// NOTE: Panic, if TypKind != Signature
func (fa *facade) IsMethod() bool {
	return fa.signature().Recv() != nil
}

// Params returns the parameters of signature s, or nil.
// NOTE: Panic, if TypKind != Signature
func (fa *facade) Params() *types.Tuple {
	return fa.signature().Params()
}

// Recv returns the receiver of signature s (if a method), or nil if a
// function. It is ignored when comparing signatures for identity.
//
// For an abstract method, Recv returns the enclosing interface either
// as a *Named or an *Interface. Due to embedding, an interface may
// contain methods whose receiver type is a different interface.
// NOTE: Panic, if TypKind != Signature
func (fa *facade) Recv() *types.Var {
	return fa.signature().Recv()
}

// Results returns the results of signature s, or nil.
// NOTE: Panic, if TypKind != Signature
func (fa *facade) Results() *types.Tuple {
	return fa.signature().Results()
}

// Variadic reports whether the signature s is variadic.
// NOTE: Panic, if TypKind != Signature
func (fa *facade) Variadic() bool {
	return fa.signature().Variadic()
}
