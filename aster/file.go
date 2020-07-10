package aster

import (
	"go/ast"
	"go/token"
	"strings"
)

// File the 'ast.File' with filename and fileSet.
type File struct {
	Filename string
	*ast.File
	*PackageInfo
	facade []*facade
}

// Overwrite originImportPath with importPath
func (f *File) CoverImport(originImportPath string, importPath string, alias ...string) {
	originImportPath = validPkgPath(originImportPath)
	importPath = validPkgPath(importPath)
	for _, im := range f.Imports {
		if im.Path.Value == originImportPath {
			im.Path.Value = importPath
			if len(alias) > 0 {
				im.Name.Name = alias[0]
			}
		}
	}
}

// Add a new import package
func (f *File) AddImport(importPath string, alias ...string) {
	importPath = validPkgPath(importPath)
	var newImport *ast.ImportSpec
	if len(alias) > 0 {
		newImport = &ast.ImportSpec{
			Name: &ast.Ident{Name: alias[0]},
			Path: &ast.BasicLit{Kind: token.STRING, Value: importPath},
		}
	} else {
		newImport = &ast.ImportSpec{
			Path: &ast.BasicLit{Kind: token.STRING, Value: importPath},
		}
	}
	f.Imports = append(f.Imports, newImport)
	for i, decl := range f.Decls {
		d, ok := decl.(*ast.GenDecl)
		if ok {
			f.Decls[i].(*ast.GenDecl).Specs = append(d.Specs, newImport)
			break
		}
	}
}

func validPkgPath(path string) string {
	if !strings.HasPrefix(path, `"`) {
		path = `"` + path
	}
	if !strings.HasSuffix(path, `"`) {
		path = path + `"`
	}
	return path
}
