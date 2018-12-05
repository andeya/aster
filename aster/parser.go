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
	"bytes"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

// ParseDir calls ParseFile for all files with names ending in ".go" in the
// directory specified by path and returns a map of package name -> package
// AST with all the packages found.
//
// If filter != nil, only the files with os.FileInfo entries passing through
// the filter (and ending in ".go") are considered. The mode bits are passed
// to ParseFile unchanged. Position information is recorded in fset, which
// must not be nil.
//
// If the directory couldn't be read, a nil map and the respective error are
// returned. If a parse error occurred, a non-nil but incomplete map and the
// first error encountered are returned.
//
func ParseDir(dir string, filter func(os.FileInfo) bool, mode ...parser.Mode) (module *Module, first error) {
	module = &Module{
		FileSet: token.NewFileSet(),
		Dir:     dir,
		filter:  filter,
		mode:    parser.ParseComments,
	}
	for _, m := range mode {
		module.mode |= m
	}
	first = module.Reparse()
	return
}

// Reparse reparses AST.
func (m *Module) Reparse() (first error) {
	pkgs, first := parser.ParseDir(m.FileSet, m.Dir, m.filter, m.mode)
	if first != nil {
		return
	}
	m.Packages = make(map[string]*Package, len(pkgs))
	for k, v := range pkgs {
		m.Packages[k] = convertPackage(m, k, v)
	}
	return
}

// ParseFile parses the source code of a single Go source file and returns
// the corresponding ast.File node. The source code may be provided via
// the filename of the source file, or via the src parameter.
//
// If src != nil, ParseFile parses the source from src and the filename is
// only used when recording position information. The type of the argument
// for the src parameter must be string, []byte, or io.Reader.
// If src == nil, ParseFile parses the file specified by filename.
//
// The mode parameter controls the amount of source text parsed and other
// optional parser functionality. Position information is recorded in the
// file set fset, which must not be nil.
//
// If the source couldn't be read, the returned AST is nil and the error
// indicates the specific failure. If the source was read but syntax
// errors were found, the result is a partial AST (with ast.Bad* objects
// representing the fragments of erroneous source code). Multiple errors
// are returned via a scanner.ErrorList which is sorted by file position.
//
func ParseFile(filename string, src interface{}, mode ...parser.Mode) (f *File, err error) {
	b, err := readSource(filename, src)
	if err != nil {
		return nil, err
	}
	f = &File{
		FileSet:  token.NewFileSet(),
		Filename: filename,
		Src:      b,
		mode:     parser.ParseComments,
	}
	for _, m := range mode {
		f.mode |= m
	}
	err = f.Reparse()
	return
}

// Reparse reparses AST.
func (f *File) Reparse() (err error) {
	b, err := readSource(f.Filename, f.Src)
	if err != nil {
		return err
	}
	f.Src = b
	file, err := parser.ParseFile(f.FileSet, f.Filename, b, f.mode)
	if err != nil {
		return
	}
	f.File = file
	if file.Name != nil {
		f.PkgName = file.Name.Name
	}
	f.setImports()
	f.collectObjects(true)
	return
}

func (f *File) setImports() {
	for _, v := range f.File.Imports {
		imp := &Import{
			ImportSpec: v,
			Path:       v.Path.Value[1 : len(v.Path.Value)-1],
			Doc:        v.Doc,
		}
		if v.Name != nil {
			imp.Name = v.Name.Name
		} else {
			imp.Name = imp.Path[strings.LastIndex(imp.Path, "/")+1:]
		}
		f.Imports = append(f.Imports, imp)
	}
}

func convertPackage(mod *Module, dir string, pkg *ast.Package) *Package {
	p := &Package{
		FileSet: mod.FileSet,
		Dir:     dir,
		Name:    pkg.Name,
		Scope:   pkg.Scope,
		Imports: pkg.Imports,
		mode:    mod.mode,
		module:  mod,
	}
	p.Files = make(map[string]*File, len(pkg.Files))
	for k, v := range pkg.Files {
		p.Files[k] = convertFile(p, k, v)
	}
	p.collectObjects()
	return p
}

func convertFile(pkg *Package, filename string, file *ast.File) *File {
	b, _ := readSource(filename, nil)
	f := &File{
		FileSet:  pkg.FileSet,
		Filename: filename,
		PkgName:  pkg.Name,
		Src:      b,
		File:     file,
		mode:     pkg.mode,
		pkg:      pkg,
	}
	f.setImports()
	return f
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

func isVariadic(t *ast.FuncType) bool {
	params := t.Params
	if num := len(params.List); num > 0 {
		_, ok := params.List[num-1].Type.(*ast.Ellipsis)
		if ok {
			return true
		}
	}
	return false
}
