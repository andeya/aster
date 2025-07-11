// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package aster

import (
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"unicode"

	"github.com/andeya/aster/internal/astutil"
)

func cloneIdent(i *ast.Ident) *ast.Ident {
	return &ast.Ident{
		Name: i.Name,
		Obj:  i.Obj,
	}
}

func cloneBasicLit(b *ast.BasicLit) *ast.BasicLit {
	if b == nil {
		return nil
	}
	return &ast.BasicLit{
		Kind:  b.Kind,
		Value: b.Value,
	}
}

func textOrError(text string, err error) string {
	if err == nil {
		return text
	}
	return "// aster: " + err.Error()
}

var filenameID int32

func autoFilename(f *ast.File) string {
	id := strconv.Itoa(int(atomic.AddInt32(&filenameID, 1)))
	return f.Name.Name + "/" + id + ".go"
}

func objectKind(obj types.Object) string {
	switch obj := obj.(type) {
	case *types.PkgName:
		return "imported package name"
	case *types.TypeName:
		return "type"
	case *types.Var:
		if obj.IsField() {
			return "field"
		}
	case *types.Func:
		if obj.Type().(*types.Signature).Recv() != nil {
			return "method"
		}
	}
	// label, func, var, const
	return strings.ToLower(strings.TrimPrefix(reflect.TypeOf(obj).String(), "*types."))
}

func typeKind(T types.Type) string {
	return strings.ToLower(strings.TrimPrefix(reflect.TypeOf(T.Underlying()).String(), "*types."))
}

// NB: for renamings, blank is not considered valid.
func isValidIdentifier(id string) bool {
	if id == "" || id == "_" {
		return false
	}
	for i, r := range id {
		if !isLetter(r) && (i == 0 || !isDigit(r)) {
			return false
		}
	}
	return token.Lookup(id) == token.IDENT
}

// isLocal reports whether obj is local to some function.
// Precondition: not a struct field or interface method.
func isLocal(obj types.Object) bool {
	// [... 5=stmt 4=func 3=file 2=pkg 1=universe]
	var depth int
	for scope := obj.Parent(); scope != nil; scope = scope.Parent() {
		depth++
	}
	return depth >= 4
}

func isPackageLevel(obj types.Object) bool {
	return obj.Pkg().Scope().Lookup(obj.Name()) == obj
}

// -- Plundered from go/scanner: ---------------------------------------

func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch >= 0x80 && unicode.IsLetter(ch)
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9' || ch >= 0x80 && unicode.IsDigit(ch)
}

// -- Plundered from github.com/andeya/aster/internal/cmd/guru -----------------

// sameFile returns true if x and y have the same basename and denote
// the same file.
func sameFile(x, y string) bool {
	if runtime.GOOS == "windows" {
		x = filepath.ToSlash(x)
		y = filepath.ToSlash(y)
	}
	if x == y {
		return true
	}
	if filepath.Base(x) == filepath.Base(y) { // (optimisation)
		if xi, err := os.Stat(x); err == nil {
			if yi, err := os.Stat(y); err == nil {
				return os.SameFile(xi, yi)
			}
		}
	}
	return false
}

// unparen returns e with any enclosing parentheses stripped.
func unparen(e ast.Expr) ast.Expr { return astutil.Unparen(e) }
