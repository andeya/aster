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

	"github.com/henrylee2cn/aster/internal/astutil"
	"github.com/henrylee2cn/aster/internal/loader"
)

// PackageInfo holds the ASTs and facts derived by the type-checker
// for a single package.
//
// Not mutated once exposed via the API.
//
type PackageInfo struct {
	prog                  *Program
	Pkg                   *types.Package
	importable            bool           // true if 'import "Pkg.Path()"' would resolve to this
	transitivelyErrorFree bool           // true if Pkg and all its dependencies are free of errors
	loaderFiles           []*loader.File // syntax trees for the package's loaderFiles
	Errors                []error        // non-nil if the package had errors
	info                  types.Info     // type-checker deductions.
	Files                 []*File
}

// newPackageInfo creates a package info.
func newPackageInfo(prog *Program, pkg *loader.PackageInfo) *PackageInfo {
	pkgInfo := &PackageInfo{
		Pkg:                   pkg.Pkg,
		importable:            pkg.Importable,
		transitivelyErrorFree: pkg.TransitivelyErrorFree,
		loaderFiles:           pkg.Files,
		Errors:                pkg.Errors,
		info:                  pkg.Info,
		prog:                  prog,
	}
	files := make([]*File, 0)
	if pkg.Files != nil {
		for _, f := range pkg.Files {
			files = append(files, &File{
				Filename:    f.Filename,
				File:        f.File,
				PackageInfo: pkgInfo,
				facade:      make([]*facade, 0),
			})
		}
	}
	pkgInfo.Files = files
	return pkgInfo
}

// Program returns the program.
func (p *PackageInfo) Program() *Program {
	return p.prog
}

// PackageInfo returns the package path.
func (p *PackageInfo) String() string {
	return p.Pkg.Path()
}

// pathEnclosingInterval returns the PackageInfo and ast.Node that
// contain source interval [start, end), and all the node's ancestors
// up to the AST root.  It searches all ast.loaderFiles in the package.
// exact is defined as for astutil.PathEnclosingInterval.
//
// The zero value is returned if not found.
//
func (p *PackageInfo) pathEnclosingInterval(start, end token.Pos) (file *loader.File, path []ast.Node, exact bool) {
	for _, f := range p.loaderFiles {
		if f.Pos() == token.NoPos {
			// This can happen if the parser saw
			// too many errors and bailed out.
			// (Use parser.AllErrors to prevent that.)
			continue
		}
		if !tokenFileContainsPos(p.prog.fset.File(f.Pos()), start) {
			continue
		}
		if path, exact := astutil.PathEnclosingInterval(f.File, start, end); path != nil {
			return f, path, exact
		}
	}
	return nil, nil, false
}

// docComment returns the doc for an identifier.
func (p *PackageInfo) docComment(id *ast.Ident) *ast.CommentGroup {
	_, nodes, _ := p.pathEnclosingInterval(id.Pos(), id.End())
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

// Preview previews the formated code and comment.
func (p *PackageInfo) Preview(ident *ast.Ident) string {
	_, nodes, _ := p.pathEnclosingInterval(ident.Pos(), ident.End())
	for _, node := range nodes {
		switch decl := node.(type) {
		case *ast.FuncDecl, *ast.GenDecl, *ast.AssignStmt:
			return textOrError(p.FormatNode(decl))
		case *ast.Field:
			s, err := p.FormatNode(decl.Type)
			if s != textOrError(s, err) {
				return s
			}
			var doc = decl.Doc.Text()
			if doc != "" {
				doc = "// " + doc
			}
			var name = decl.Names[0].Name
			return "//aster:field\n" + doc + "var " + name + " " + s
		case *ast.File:
			return "package " + ident.String()
		}
	}
	return "// aster: can not preview " + ident.String()
}
