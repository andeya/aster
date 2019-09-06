// Package aster is golang coding efficiency engine.
//
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
	"errors"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"

	"github.com/henrylee2cn/aster/internal/loader"
	"github.com/henrylee2cn/aster/tools"
	"github.com/henrylee2cn/goutil"
)

// A Program is a Go program loaded from source.
type Program struct {
	// initial
	conf         loader.Config
	initialError error // first error for initial
	initiated    bool

	// fset the file set for this program
	fset *token.FileSet

	// created[i] contains the initial package whose ASTs or
	// filenames were supplied by AddFiles(), MustAddFiles()
	// and LoadFile() followed by the external test package,
	// if any, of each package in Import(), ImportWithTests(),
	// LoadPkgs and LoadPkgsWithTests() ordered by ImportPath.
	//
	// NOTE: these files must not import "C".  Cgo preprocessing is
	// only performed on imported packages, not ad hoc packages.
	//
	created []*PackageInfo

	// imported contains the initially imported packages,
	// as specified by Import(), ImportWithTests(), LoadPkgs and LoadPkgsWithTests().
	imported map[string]*PackageInfo

	// allPackages contains the PackageInfo of every package
	// encountered by Load: all initial packages and all
	// dependencies, including incomplete ones.
	allPackages map[*types.Package]*PackageInfo

	// We use token.File, not filename, since a file may appear to
	// belong to multiple packages and be parsed more than once.
	// token.File captures this distinction; filename does not.
	filesToUpdate map[*token.File]bool
	// <filename, codes> Non file Sources
	nonfileSources map[string][]byte
}

// LoadFile parses the source code of a single Go file and loads a new program.
//
// src specifies the parser input as a string, []byte, or io.Reader, and
// filename is its apparent name.  If src is nil, the contents of
// filename are read from the file system.
//
func LoadFile(filename string, src interface{}) (*Program, error) {
	return NewProgram().AddFile(filename, src).Load()
}

// LoadDirs parses the source code of Go files under the directories and loads a new program.
func LoadDirs(dirs ...string) (*Program, error) {
	p := NewProgram()
	srcs, _ := goutil.StringsConvert(build.Default.SrcDirs(), func(s string) (string, error) {
		return s + string(filepath.Separator), nil
	})
	for _, dir := range dirs {
		if !filepath.IsAbs(dir) {
			dir, _ = filepath.Abs(dir)
		}
		err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
			if err != nil || !f.IsDir() {
				return nil
			}
			for _, src := range srcs {
				pkg := strings.TrimPrefix(path, src)
				if pkg != path {
					p.Import(pkg)
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return p.Load()
}

// LoadPkgs imports packages and loads a new program.
//
// the set of initial source packages located relative to $GOPATH.
//
func LoadPkgs(pkgPath ...string) (*Program, error) {
	return NewProgram().Import(pkgPath...).Load()
}

// LoadPkgsWithTests imports packages and loads a new program.
//
// the set of initial source packages located relative to $GOPATH.
//
// The package will be augmented by any *_test.go files in
// its directory that contain a "package x" (not "package x_test")
// declaration.
//
// In addition, if any *_test.go files contain a "package x_test"
// declaration, an additional package comprising just those files will
// be added to CreatePkgs.
//
func LoadPkgsWithTests(pkgPath ...string) (*Program, error) {
	return NewProgram().ImportWithTests(pkgPath...).Load()
}

// NewProgram creates a empty program.
func NewProgram() *Program {
	prog := new(Program)
	prog.filesToUpdate = make(map[*token.File]bool, 128)
	prog.conf.ParserMode = parser.ParseComments | parser.AllErrors
	// Optimization: don't type-check the bodies of functions in our
	// dependencies, since we only need exported package members.
	prog.conf.TypeCheckFuncBodies = func(p string) bool {
		pp := strings.TrimSuffix(p, "_test")
		for _, cp := range prog.conf.CreatePkgs {
			if cp.Path == p || cp.Path == pp {
				return true
			}
		}
		for k := range prog.conf.ImportPkgs {
			if k == p || k == pp {
				return true
			}
		}
		return false
	}
	// Ideally we would just return conf.Load() here, but go/types
	// reports certain "soft" errors that gc does not (Go issue 14596).
	// As a workaround, we set AllowErrors=true and then duplicate
	// the loader's error checking but allow soft errors.
	// It would be nice if the loader API permitted "AllowErrors: soft".
	prog.conf.AllowErrors = true
	prog.conf.TypeChecker.DisableUnusedImportCheck = true
	prog.nonfileSources = make(map[string][]byte)
	return prog
}

// AddFile parses the source code of a single Go source file.
//
// src specifies the parser input as a string, []byte, or io.Reader, and
// filename is its apparent name.  If src is nil, the contents of
// filename are read from the file system.
//
// filename is used to rewrite to local file;
// if empty, rewrite to self-increasing number filename under the package name path.
//
func (prog *Program) AddFile(filename string, src interface{}) (itself *Program) {
	if !prog.initiated && prog.initialError == nil {
		b, srcErr := tools.ReadSourceBytes(src)
		f, err := prog.conf.ParseFile(filename, b)
		if err != nil {
			prog.initialError = err
		} else {
			if filename == "" {
				filename = autoFilename(f)
			}
			prog.conf.CreateFromFiles(f.Name.Name, &loader.File{Filename: filename, File: f})
			if srcErr == nil {
				prog.nonfileSources[filename] = b
			}
		}
	}
	return prog
}

// Import imports packages that will be imported from source,
// the set of initial source packages located relative to $GOPATH.
func (prog *Program) Import(pkgPath ...string) (itself *Program) {
	if !prog.initiated && prog.initialError == nil {
		for _, p := range pkgPath {
			prog.conf.Import(p)
		}
	}
	return prog
}

// ImportWithTests imports packages that will be imported from source,
// the set of initial source packages located relative to $GOPATH.
// The package will be augmented by any *_test.go files in
// its directory that contain a "package x" (not "package x_test")
// declaration.
//
// In addition, if any *_test.go files contain a "package x_test"
// declaration, an additional package comprising just those files will
// be added to CreatePkgs.
//
func (prog *Program) ImportWithTests(pkgPath ...string) (itself *Program) {
	if !prog.initiated && prog.initialError == nil {
		for _, p := range pkgPath {
			prog.conf.ImportWithTests(p)
		}
	}
	return prog
}

// Load loads the program's packages,
// and loads their dependencies packages as needed.
//
// On failure, returns an error.
// It is an error if no packages were loaded.
//
func (prog *Program) Load() (itself *Program, err error) {
	if prog.initiated {
		return prog, errors.New("can not load two times")
	}
	if prog.initialError != nil {
		return prog, prog.initialError
	}
	prog.initiated = true
	defer func() {
		if p := recover(); p != nil {
			prog.initialError = fmt.Errorf("%v", p)
		}
	}()
	p, err := prog.conf.Load()
	if err != nil {
		prog.initialError = err
		return prog, prog.initialError
	}
	var errpkgs []string
	// Report hard errors in indirectly imported packages.
	for _, info := range p.AllPackages {
		if containsHardErrors(info.Errors) {
			errpkgs = append(errpkgs, info.Pkg.Path())
		}
	}
	if errpkgs != nil {
		var more string
		if len(errpkgs) > 3 {
			more = fmt.Sprintf(" and %d more", len(errpkgs)-3)
			errpkgs = errpkgs[:3]
		}
		prog.initialError = fmt.Errorf("couldn't load packages due to errors: %s%s",
			strings.Join(errpkgs, ", "), more)
		return prog, prog.initialError
	}
	return prog.convert(p), prog.initialError
}

// MustLoad is the same as Load(), but panic when error occur.
func (prog *Program) MustLoad() (itself *Program) {
	_, err := prog.Load()
	if err != nil {
		panic(err)
	}
	return prog
}

func (prog *Program) convert(p *loader.Program) (itself *Program) {
	prog.fset = p.Fset
	prog.imported = make(map[string]*PackageInfo, len(prog.imported))
	prog.allPackages = make(map[*types.Package]*PackageInfo, len(prog.allPackages))

	var solved = make(map[*loader.PackageInfo]*PackageInfo, len(p.AllPackages))
	for _, info := range p.Created {
		newInfo := newPackageInfo(prog, info)
		solved[info] = newInfo
		prog.created = append(prog.created, newInfo)
	}
	for k, info := range p.Imported {
		newInfo := newPackageInfo(prog, info)
		solved[info] = newInfo
		prog.imported[k] = newInfo
	}
	for k, info := range p.AllPackages {
		if newInfo, ok := solved[info]; ok {
			prog.allPackages[k] = newInfo
		} else {
			newInfo := newPackageInfo(prog, info)
			solved[info] = newInfo
			prog.allPackages[k] = newInfo
		}
	}
	prog.check()
	return prog
}

func (prog *Program) check() {
	for _, pkg := range prog.InitialPackages() {
		pkg.check()
	}
}

// InitialPackages returns a new slice containing the set of initial
// packages (created + imported) in unspecified order.
func (prog *Program) InitialPackages() []*PackageInfo {
	pkgs := make([]*PackageInfo, 0, len(prog.created)+len(prog.imported))
	pkgs = append(pkgs, prog.created...)
	for _, pkg := range prog.imported {
		pkgs = append(pkgs, pkg)
	}
	return pkgs
}

// Package returns the ASTs and results of type checking for the
// specified package.
// NOTE: return nil, if the package does not exist.
func (prog *Program) Package(path string) *PackageInfo {
	for k, v := range prog.allPackages {
		if k.Path() == path {
			return v
		}
	}
	for _, info := range prog.created {
		if path == info.Pkg.Path() {
			return info
		}
	}
	return nil
}

// pathEnclosingInterval returns the PackageInfo and ast.Node that
// contain source interval [start, end), and all the node's ancestors
// up to the AST root.  It searches all ast.Files of all packages in prog.
// exact is defined as for astutil.PathEnclosingInterval.
//
// The zero value is returned if not found.
//
func (prog *Program) pathEnclosingInterval(start, end token.Pos) (pkg *PackageInfo, file *loader.File, path []ast.Node, exact bool) {
	for _, pkg = range prog.allPackages {
		file, path, exact = pkg.pathEnclosingInterval(start, end)
		if path != nil {
			return
		}
	}
	return nil, nil, nil, false
}

func (prog *Program) source(filename string) ([]byte, error) {
	src, ok := prog.nonfileSources[filename]
	if ok {
		return src, nil
	}
	return tools.ReadSource(filename, nil)
}

func containsHardErrors(errors []error) bool {
	for _, err := range errors {
		if err, ok := err.(types.Error); ok && err.Soft {
			continue
		}
		return true
	}
	return false
}

func tokenFileContainsPos(f *token.File, pos token.Pos) bool {
	p := int(pos)
	base := f.Base()
	return base <= p && p < base+f.Size()
}
