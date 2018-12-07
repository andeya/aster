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
	"go/ast"
	"go/format"
	"os"
	"path/filepath"

	"github.com/henrylee2cn/goutil"
)

// Format format the package and returns the string.
// @codes <fileName,code>
func (p *PackageInfo) Format() (codes map[string]string, first error) {
	codes = make(map[string]string, len(p.Files))
	var code string
	for _, f := range p.Files {
		code, first = p.FormatNode(f)
		if first != nil {
			return
		}
		codes[f.Name.String()] = code
	}
	return
}

// FormatNode formats the node and returns the string.
func (p *PackageInfo) FormatNode(node ast.Node) (string, error) {
	var dst bytes.Buffer
	err := format.Node(&dst, p.prog.Fset, node)
	if err != nil {
		return "", err
	}
	return goutil.BytesToString(dst.Bytes()), nil
}

// Rewrite formats the package codes and writes to the local files.
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
