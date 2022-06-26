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

// Elem returns the element type.
// NOTE: Panic, if TypKind != (Array, Slice, Map, Chan and Pointer)
func (fa *facade) Elem() types.Type {
	typ := fa.typ()
	switch t := typ.(type) {
	default:
		panic(fmt.Sprintf("aster: Elem of non-TypKind(Array, Slice, Map, Chan and Pointer): %T", typ))
	case *types.Array:
		return t.Elem()
	case *types.Slice:
		return t.Elem()
	case *types.Map:
		return t.Elem()
	case *types.Chan:
		return t.Elem()
	case *types.Pointer:
		return t.Elem()
	}
}

// NOTE: Panic, if TypKind != Map
func (fa *facade) dict() *types.Map {
	typ := fa.typ()
	t, ok := typ.(*types.Map)
	if !ok {
		panic(fmt.Sprintf("aster: dict of non-Map TypKind: %T", typ))
	}
	return t
}

// Key returns the key type of map.
// NOTE: Panic, if TypKind != Map
func (fa *facade) Key() types.Type {
	return fa.dict().Key()
}

// NOTE: Panic, if TypKind != Array
func (fa *facade) array() *types.Array {
	typ := fa.typ()
	t, ok := typ.(*types.Array)
	if !ok {
		panic(fmt.Sprintf("aster: array of non-Array TypKind: %T", typ))
	}
	return t
}

// Len returns the length of array, or the number variables of tuple.
// A negative result indicates an unknown length.
// NOTE: Panic, if TypKind != Array and TypKind != Tuple
func (fa *facade) Len() int64 {
	typ := fa.typ()
	switch t := typ.(type) {
	default:
		panic(fmt.Sprintf("aster: Elem of non-(Array, Slice, Map, Chan and Pointer) TypKind: %T", typ))
	case *types.Array:
		return t.Len()
	case *types.Tuple:
		return int64(t.Len())
	}
}

// NOTE: Panic, if TypKind != Chan
func (fa *facade) channle() *types.Chan {
	typ := fa.typ()
	t, ok := typ.(*types.Chan)
	if !ok {
		panic(fmt.Sprintf("aster: channle of non-Chan TypKind: %T", typ))
	}
	return t
}

// ChanDir returns the direction of channel.
// NOTE: Panic, if TypKind != Chan
func (fa *facade) ChanDir() types.ChanDir {
	return fa.channle().Dir()
}

// NOTE: Panic, if TypKind != Basic
func (fa *facade) basic() *types.Basic {
	typ := fa.typ()
	t, ok := typ.(*types.Basic)
	if !ok {
		panic(fmt.Sprintf("aster: basic of non-Basic TypKind: %T", typ))
	}
	return t
}

// BasicInfo returns information about properties of basic type.
// NOTE: Panic, if TypKind != Basic
func (fa *facade) BasicInfo() types.BasicInfo {
	return fa.basic().Info()
}

// BasicKind returns the kind of basic type.
// NOTE: Panic, if TypKind != Basic
func (fa *facade) BasicKind() types.BasicKind {
	return fa.basic().Kind()
}
