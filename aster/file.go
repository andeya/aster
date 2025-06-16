package aster

import (
	"errors"
	"go/ast"
	"go/token"
	"strings"
)

// File the 'ast.File' with filename and fileSet.
type File struct {
	Filename string
	*ast.File
	*PackageInfo
	facades []*facade
}

// FindImportByPath find import by import path, and return alias and found result.
func (f *File) FindImportByPath(importPath string) (alias string, found bool) {
	importPath = validPkgPath(importPath)
	for _, im := range f.Imports {
		if im.Path.Value == importPath {
			if im.Name != nil {
				return im.Name.Name, true
			}
			return "", true
		}
	}
	return "", false
}

// FindImportAlias find import alias by import path, and return alias and found result.
func (f *File) FindImportAlias(alias string) (importPath string, found bool) {
	for _, im := range f.Imports {
		if im.Name != nil && im.Name.Name == alias {
			return im.Path.Value, true
		}
	}
	return "", false
}

// CoverImport cover originImportPath with importPath
func (f *File) CoverImport(originImportPath string, importPath string, alias ...string) {
	originImportPath = validPkgPath(originImportPath)
	importPath = validPkgPath(importPath)
	for _, im := range f.Imports {
		if im.Path.Value == originImportPath {
			im.Path.Value = importPath
			if len(alias) > 0 {
				im.Name = &ast.Ident{Name: alias[0]}
			}
		}
	}
}

// AddImport add a new import package
func (f *File) AddImport(importPath string, alias ...string) error {
	importPath = validPkgPath(importPath)
	for _, im := range f.Imports {
		if im.Path.Value == importPath || (len(alias) > 0 && im.Name != nil && im.Name.Name == alias[0]) {
			return errors.New("add import package or alias is exist")
		}
	}
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
	if f.Decls == nil || len(f.Decls) == 0 {
		f.Decls = []ast.Decl{&ast.GenDecl{Tok: token.IMPORT}}
	}
	for i, decl := range f.Decls {
		d, ok := decl.(*ast.GenDecl)
		if ok && d.Tok == token.IMPORT {
			f.Decls[i].(*ast.GenDecl).Specs = append(d.Specs, newImport)
			break
		}
	}
	return nil
}

// DelImport delete a import path
func (f *File) DelImport(path string) {
	path = validPkgPath(path)
	var delIm *ast.ImportSpec
	for i, im := range f.Imports {
		if im.Path.Value == path {
			delIm = im
			f.Imports = append(f.Imports[0:i], f.Imports[i+1:]...)
			break
		}
	}
	if delIm == nil {
		return
	}
	for _, decl := range f.Decls {
		d, ok := decl.(*ast.GenDecl)
		if ok && d.Tok == token.IMPORT {
			for i, s := range d.Specs {
				if s == delIm {
					f.Decls[i].(*ast.GenDecl).Specs = append(d.Specs[0:i], d.Specs[i+1:]...)
				}
			}
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
