// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package aster

import (
	"bytes"
	"errors"
	"go/ast"
	"go/token"
	"go/types"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"unicode"

	"github.com/henrylee2cn/goutil"
	"github.com/henrylee2cn/aster/aster/internal/astutil"
)

var pkglineRegexp = regexp.MustCompile("\n*package[\t ]+([^/\n]+)[/\n]")

// ChangePkgName change package name of the code and return the new code.
func ChangePkgName(code string, pkgname string) string {
	s := strings.TrimSpace(pkglineRegexp.FindString(code))
	s = strings.TrimSpace(strings.TrimRight(s, "/"))
	if s == "" {
		return code
	}
	return strings.Replace(code, s, "package "+pkgname, 1)
}

// PkgName get the package name of the code, file or directory.
// NOTE:
//  If src==nil, find the package name from the file or directory specified by 'filenameOrDirectory';
//  If src!=nil, find the package name from the code represented by 'src'.
func PkgName(filenameOrDirectory string, src interface{}) (string, error) {
	if src == nil {
		existed, isDir := goutil.FileExist(filenameOrDirectory)
		if !existed {
			return "", errors.New("file or directory is not existed")
		}
		if isDir {
			err := filepath.Walk(filenameOrDirectory, func(path string, f os.FileInfo, err error) error {
				if err != nil || f.IsDir() {
					return nil
				}
				if strings.HasSuffix(path, ".go") {
					filenameOrDirectory = path
					return errors.New("")
				}
				return nil
			})
			if err == nil || err.Error() != "" {
				return "", err
			}
		}
	}
	b, err := readSource(filenameOrDirectory, src)
	if err != nil {
		return "", err
	}
	r := pkglineRegexp.FindSubmatch(b)
	if len(r) < 2 {
		return "", nil
	}
	return goutil.BytesToString(bytes.TrimSpace(r[1])), nil
}

func readSource(filename string, src interface{}) ([]byte, error) {
	if src != nil {
		switch s := src.(type) {
		case string:
			return []byte(s), nil
		case []byte:
			return s, nil
		case *bytes.Buffer:
			// is io.Reader, but src is already available in []byte form
			if s != nil {
				return s.Bytes(), nil
			}
		case io.Reader:
			return ioutil.ReadAll(s)
		}
		return nil, errors.New("invalid source")
	}
	return ioutil.ReadFile(filename)
}

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

// -- Plundered from github.com/henrylee2cn/aster/aster/internal/cmd/guru -----------------

// sameFile returns true if x and y have the same basename and denote
// the same file.
//
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
