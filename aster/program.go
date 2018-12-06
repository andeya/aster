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
	"go/parser"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/loader"
)

// A Program is a Go program loaded from source.
type Program struct {
	// initial
	conf         loader.Config
	initialError error // first error for initial
	initiated    bool

	// Fset the file set for this program
	Fset *token.FileSet

	// Created[i] contains the initial package whose ASTs or
	// filenames were supplied by AddFiles() and MustAddFiles(), followed by
	// the external test package, if any, of each package in
	// Import() and ImportWithTests() ordered by ImportPath.
	//
	// NOTE: these files must not import "C".  Cgo preprocessing is
	// only performed on imported packages, not ad hoc packages.
	//
	// TODO(adonovan): we need to copy and adapt the logic of
	// goFilesPackage (from $GOROOT/src/cmd/go/build.go) and make
	// Config.Import and Config.Create methods return the same kind
	// of entity, essentially a build.Package.
	// Perhaps we can even reuse that type directly.
	Created []*PackageInfo

	// Imported contains the initially imported packages,
	// as specified by Import() and ImportWithTests().
	Imported map[string]*PackageInfo

	// AllPackages contains the PackageInfo of every package
	// encountered by Load: all initial packages and all
	// dependencies, including incomplete ones.
	AllPackages map[*types.Package]*PackageInfo
}

// PackageInfo holds the ASTs and facts derived by the type-checker
// for a single package.
//
// Not mutated once exposed via the API.
//
type PackageInfo struct {
	Pkg                   *types.Package
	Importable            bool        // true if 'import "Pkg.Path()"' would resolve to this
	TransitivelyErrorFree bool        // true if Pkg and all its dependencies are free of errors
	Files                 []*ast.File // syntax trees for the package's files
	Errors                []error     // non-nil if the package had errors
	types.Info                        // type-checker deductions.
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
	prog.conf.ParserMode = parser.ParseComments
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
	return prog
}

// AddFile parses the source code of a single Go source file.
//
// src specifies the parser input as a string, []byte, or io.Reader, and
// filename is its apparent name.  If src is nil, the contents of
// filename are read from the file system.
//
func (prog *Program) AddFile(filename string, src interface{}) (itself *Program) {
	if !prog.initiated && prog.initialError == nil {
		f, err := prog.conf.ParseFile(filename, src)
		if err != nil {
			prog.initialError = err
		} else {
			prog.conf.CreateFromFiles(f.Name.Name, f)
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
	return prog.bind(p), prog.initialError
}

// MustLoad is the same as Load(), but panic when error occur.
func (prog *Program) MustLoad() (itself *Program) {
	_, err := prog.Load()
	if err != nil {
		panic(err)
	}
	return prog
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

func (prog *Program) bind(p *loader.Program) (itself *Program) {
	prog.Fset = p.Fset
	prog.Imported = make(map[string]*PackageInfo, len(prog.Imported))
	prog.AllPackages = make(map[*types.Package]*PackageInfo, len(prog.AllPackages))

	var solved = make(map[*loader.PackageInfo]*PackageInfo, len(p.AllPackages))
	for _, info := range p.Created {
		newInfo := newPackageInfo(info)
		solved[info] = newInfo
		prog.Created = append(prog.Created, newInfo)
	}
	for k, info := range p.Imported {
		newInfo := newPackageInfo(info)
		solved[info] = newInfo
		prog.Imported[k] = newInfo
	}
	for k, info := range p.AllPackages {
		if newInfo, ok := solved[info]; ok {
			prog.AllPackages[k] = newInfo
		} else {
			newInfo := newPackageInfo(info)
			solved[info] = newInfo
			prog.AllPackages[k] = newInfo
		}
	}
	return prog
}

// newPackageInfo creates a package info.
func newPackageInfo(pkg *loader.PackageInfo) *PackageInfo {
	return &PackageInfo{
		Pkg:                   pkg.Pkg,
		Importable:            pkg.Importable,
		TransitivelyErrorFree: pkg.TransitivelyErrorFree,
		Files:                 pkg.Files,
		Errors:                pkg.Errors,
		Info:                  pkg.Info,
	}
}

// InitialPackages returns a new slice containing the set of initial
// packages (Created + Imported) in unspecified order.
func (prog *Program) InitialPackages() []*PackageInfo {
	infos := make([]*PackageInfo, 0, len(prog.Created)+len(prog.Imported))
	infos = append(infos, prog.Created...)
	for _, info := range prog.Imported {
		infos = append(infos, info)
	}
	return infos
}

// Package returns the ASTs and results of type checking for the
// specified package.
// NOTE: return nil, if the package does not exist.
func (prog *Program) Package(path string) *PackageInfo {
	for k, v := range prog.AllPackages {
		if k.Path() == path {
			return v
		}
	}
	for _, info := range prog.Created {
		if path == info.Pkg.Path() {
			return info
		}
	}
	return nil
}

func (p *PackageInfo) String() string {
	return p.Pkg.Path()
}
