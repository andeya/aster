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
	// Fset the file set for this program
	Fset *token.FileSet

	// Created[i] contains the initial package whose ASTs or
	// filenames were supplied by AddFiles() and MustAddFiles(), followed by
	// the external test package, if any, of each package in
	// Imports() and ImportWithTests() ordered by ImportPath.
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
	// as specified by Imports() and ImportWithTests().
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

// Loader loading a whole program from Go source code.
type Loader struct {
	conf loader.Config
}

// NewLoader creates a loader.
func NewLoader(mode ...parser.Mode) *Loader {
	loader := new(Loader)
	loader.conf.ParserMode = parser.ParseComments
	for _, m := range mode {
		loader.conf.ParserMode |= m
	}
	// Optimization: don't type-check the bodies of functions in our
	// dependencies, since we only need exported package members.
	loader.conf.TypeCheckFuncBodies = func(p string) bool {
		pp := strings.TrimSuffix(p, "_test")
		for _, cp := range loader.conf.CreatePkgs {
			if cp.Path == p || cp.Path == pp {
				return true
			}
		}
		for k := range loader.conf.ImportPkgs {
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
	loader.conf.AllowErrors = true
	return loader
}

// AddFile parses the source code of a single Go source file.
//
// src specifies the parser input as a string, []byte, or io.Reader, and
// filename is its apparent name.  If src is nil, the contents of
// filename are read from the file system.
//
func (l *Loader) AddFile(filename string, src interface{}) error {
	f, err := l.conf.ParseFile(filename, src)
	if err != nil {
		return err
	}
	l.conf.CreateFromFiles(f.Name.Name, f)
	return nil
}

// MustAddFile is similar to AddFile, but panic when existing error.
func (l *Loader) MustAddFile(filename string, src interface{}) *Loader {
	err := l.AddFile(filename, src)
	if err != nil {
		panic(err)
	}
	return l
}

// Imports imports packages that will be imported from source,
// the set of initial source packages located relative to $GOPATH.
func (l *Loader) Imports(pkgPath ...string) *Loader {
	for _, p := range pkgPath {
		l.conf.Import(p)
	}
	return l
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
func (l *Loader) ImportWithTests(pkgPath ...string) *Loader {
	for _, p := range pkgPath {
		l.conf.ImportWithTests(p)
	}
	return l
}

// Load creates the initial packages specified by conf.{Create,Import}Pkgs,
// loading their dependencies packages as needed.
//
// On success, Load returns a Program containing a PackageInfo for
// each package.  On failure, it returns an error.
//
// It is an error if no packages were loaded.
//
func (l *Loader) Load() (prog *Program, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("%v", p)
		}
	}()
	p, err := l.conf.Load()
	if err != nil {
		return nil, err
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
		return nil, fmt.Errorf("couldn't load packages due to errors: %s%s",
			strings.Join(errpkgs, ", "), more)
	}
	return newProgram(p), nil
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

// MustLoad is similar to Load, but panic when existing error.
func (l *Loader) MustLoad() *Program {
	p, err := l.Load()
	if err != nil {
		panic(err)
	}
	return p
}

// newProgram creates a program object.
func newProgram(prog *loader.Program) *Program {
	p := &Program{
		Fset:        prog.Fset,
		Imported:    make(map[string]*PackageInfo, len(prog.Imported)),
		AllPackages: make(map[*types.Package]*PackageInfo, len(prog.AllPackages)),
	}
	var solved = make(map[*loader.PackageInfo]*PackageInfo, len(prog.AllPackages))
	for _, pkg := range prog.Created {
		newPkg := newPackageInfo(pkg)
		solved[pkg] = newPkg
		p.Created = append(p.Created, newPkg)
	}
	for k, pkg := range prog.Imported {
		newPkg := newPackageInfo(pkg)
		solved[pkg] = newPkg
		p.Imported[k] = newPkg
	}
	for k, pkg := range prog.AllPackages {
		if newPkg, ok := solved[pkg]; ok {
			p.AllPackages[k] = newPkg
		} else {
			newPkg := newPackageInfo(pkg)
			solved[pkg] = newPkg
			p.AllPackages[k] = newPkg
		}
	}
	return p
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

// Package returns the ASTs and results of type checking for the
// specified package.
// NOTE: return nil, if the package does not exist.
func (p *Program) Package(path string) *PackageInfo {
	for k, v := range p.AllPackages {
		if k.Path() == path {
			return v
		}
	}
	for _, info := range p.Created {
		if path == info.Pkg.Path() {
			return info
		}
	}
	return nil
}

func (p *PackageInfo) String() string {
	return p.Pkg.Path()
}

// InitialPackages returns a new slice containing the set of initial
// packages (Created + Imported) in unspecified order.
func (p *Program) InitialPackages() []*PackageInfo {
	infos := make([]*PackageInfo, 0, len(p.Created)+len(p.Imported))
	infos = append(infos, p.Created...)
	for _, info := range p.Imported {
		infos = append(infos, info)
	}
	return infos
}
