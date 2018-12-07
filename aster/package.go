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
	"go/token"
	"go/types"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/loader"
)

// PackageInfo holds the ASTs and facts derived by the type-checker
// for a single package.
//
// Not mutated once exposed via the API.
//
type PackageInfo struct {
	prog                  *Program
	Pkg                   *types.Package
	Importable            bool        // true if 'import "Pkg.Path()"' would resolve to this
	TransitivelyErrorFree bool        // true if Pkg and all its dependencies are free of errors
	Files                 []*ast.File // syntax trees for the package's files
	Errors                []error     // non-nil if the package had errors
	types.Info                        // type-checker deductions.
	objects               map[token.Pos]interface{}
}

// A File node represents a Go source file.
type File struct {
	*ast.File
	Filename string
}

// newPackageInfo creates a package info.
func newPackageInfo(prog *Program, pkg *loader.PackageInfo) *PackageInfo {
	return &PackageInfo{
		Pkg:                   pkg.Pkg,
		Importable:            pkg.Importable,
		TransitivelyErrorFree: pkg.TransitivelyErrorFree,
		Files:                 pkg.Files,
		Errors:                pkg.Errors,
		Info:                  pkg.Info,
		prog:                  prog,
		objects:               make(map[token.Pos]interface{}, 128),
	}
}

func (p *PackageInfo) String() string {
	return p.Pkg.Path()
}

// PathEnclosingInterval returns the PackageInfo and ast.Node that
// contain source interval [start, end), and all the node's ancestors
// up to the AST root.  It searches all ast.Files in the package.
// exact is defined as for astutil.PathEnclosingInterval.
//
// The zero value is returned if not found.
//
func (p *PackageInfo) PathEnclosingInterval(start, end token.Pos) (path []ast.Node, exact bool) {
	for _, f := range p.Files {
		if f.Pos() == token.NoPos {
			// This can happen if the parser saw
			// too many errors and bailed out.
			// (Use parser.AllErrors to prevent that.)
			continue
		}
		if !tokenFileContainsPos(p.prog.Fset.File(f.Pos()), start) {
			continue
		}
		if path, exact := astutil.PathEnclosingInterval(f, start, end); path != nil {
			return path, exact
		}
	}
	return nil, false
}

// DocComment returns the doc for an identifier.
func (p *PackageInfo) DocComment(id *ast.Ident) *ast.CommentGroup {
	nodes, _ := p.PathEnclosingInterval(id.Pos(), id.End())
	for _, node := range nodes {
		switch decl := node.(type) {
		case *ast.FuncDecl:
			return decl.Doc
		case *ast.Field:
			return decl.Doc
		case *ast.GenDecl:
			return decl.Doc
		// For {Type,Value}Spec, if the doc on the spec is absent,
		// search for the enclosing GenDecl
		case *ast.TypeSpec:
			if decl.Doc != nil {
				return decl.Doc
			}
		case *ast.ValueSpec:
			if decl.Doc != nil {
				return decl.Doc
			}
		case *ast.Ident:
		default:
			return nil
		}
	}
	return nil
}
