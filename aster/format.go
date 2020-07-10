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
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/henrylee2cn/aster/tools"
	"github.com/henrylee2cn/goutil"
)

// Format formats the created and imported packages, and returns the string.
// @codes <fileName,code>
func (prog *Program) Format() (codes map[string]string, first error) {
	codes = make(map[string]string, 128)
	var c map[string]string
	for _, pkg := range prog.InitialPackages() {
		c, first = pkg.Format()
		if first != nil {
			return
		}
		for k, v := range c {
			codes[k] = v
		}
	}
	return
}

// Format formats the package and returns the string.
// @codes <fileName,code>
func (p *PackageInfo) Format() (codes map[string]string, first error) {
	codes = make(map[string]string, len(p.loaderFiles))
	var code string
	var codeBytes []byte
	pkgName := p.Pkg.Name()
	for _, f := range p.loaderFiles {
		code, first = p.FormatNode(f.File)
		if first != nil {
			return
		}
		code = tools.ChangePkgName(code, pkgName)
		codeBytes, first = tools.Format("", goutil.StringToBytes(code), nil)
		if first != nil {
			return
		}
		codes[f.Filename] = goutil.BytesToString(codeBytes)
	}
	return
}

func (f *File) Format() (codes map[string]string, first error) {
	codes = make(map[string]string, 1)
	code, first := f.PackageInfo.FormatNode(f.File)
	codes[f.Filename] = code
	return codes, first
}

// FormatNode formats the node and returns the string.
func (prog *Program) FormatNode(node ast.Node) (string, error) {
	return formatNode(prog.fset, node)
}

// FormatNode formats the node and returns the string.
func (p *PackageInfo) FormatNode(node ast.Node) (string, error) {
	return p.prog.FormatNode(node)
}

// FormatNode formats the node and returns the string.
func (f *File) FormatNode(node ast.Node) (string, error) {
	return f.PackageInfo.FormatNode(node)
}

func formatNode(fset *token.FileSet, node ast.Node) (string, error) {
	var dst bytes.Buffer
	err := format.Node(&dst, fset, node)
	if err != nil {
		return "", err
	}
	return goutil.BytesToString(dst.Bytes()), nil
}

// Rewrite formats the created and imported packages codes and writes to local loaderFiles.
func (prog *Program) Rewrite() (first error) {
	for _, pkg := range prog.InitialPackages() {
		first = pkg.Rewrite()
		if first != nil {
			return
		}
	}
	return
}

// Rewrite formats the package codes and writes to local loaderFiles.
func (p *PackageInfo) Rewrite() (first error) {
	codes, first := p.Format()
	if first != nil {
		return
	}
	for k, v := range codes {
		first = writeFile(k, v)
		if first != nil {
			return first
		}
	}
	return
}

func (f *File) Rewrite() (first error) {
	codes, first := f.Format()
	if first != nil {
		return
	}
	for k, v := range codes {
		first = writeFile(k, v)
		if first != nil {
			return first
		}
	}
	return
}

// PrintResume prints the program resume.
func (prog *Program) PrintResume() {
	// Created packages are the initial packages specified by a call
	// to CreateFromFilenames or CreateFromFiles.
	var names []string
	for _, info := range prog.created {
		names = append(names, info.Pkg.Path())
	}
	fmt.Printf("created: %s\n", names)

	// Imported packages are the initial packages specified by a
	// call to Import or ImportWithTests.
	names = nil
	for _, info := range prog.imported {
		if strings.Contains(info.Pkg.Path(), "internal") {
			continue // skip, to reduce fragility
		}
		names = append(names, info.Pkg.Path())
	}
	sort.Strings(names)
	fmt.Printf("imported: %s\n", names)

	// InitialPackages contains the union of created and imported.
	names = nil
	for _, info := range prog.InitialPackages() {
		names = append(names, info.Pkg.Path())
	}
	sort.Strings(names)
	fmt.Printf("initial: %s\n", names)

	// AllPackages contains all initial packages and their dependencies.
	names = nil
	for pkg := range prog.allPackages {
		names = append(names, pkg.Path())
	}
	sort.Strings(names)
	fmt.Printf("all: %s\n", names)
}

func writeFile(filename, text string) error {
	filename, err := filepath.Abs(filename)
	if err != nil {
		return err
	}
	dir := filepath.Dir(filename)
	err = os.MkdirAll(dir, 0777)
	if err != nil {
		return err
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	_, err = f.Write(goutil.StringToBytes(text))
	return err
}
