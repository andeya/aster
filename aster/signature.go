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
	"go/types"
	"strings"

	"github.com/henrylee2cn/aster/internal/loader"
)

// ---------------------------------- TypKind = Signature (function) ----------------------------------

// NOTE: Panic, if TypKind != Signature
func (fa *facade) signature() *types.Signature {
	typ := fa.typ()
	t, ok := typ.(*types.Signature)
	if !ok {
		panic(fmt.Sprintf("aster: signature of non-Signature TypKind: %T", typ))
	}
	return t
}

// IsMethod returns whether it is a method.
func (fa *facade) IsMethod() bool {
	if fa.typKind() != Signature {
		return false
	}
	return fa.signature().Recv() != nil
}

// Params returns the parameters of signature s, or nil.
// NOTE: Panic, if TypKind != Signature
func (fa *facade) Params() *types.Tuple {
	return fa.signature().Params()
}

// Recv returns the receiver of signature s (if a method), or nil if a
// function. It is ignored when comparing signatures for identity.
//
// For an abstract method, Recv returns the enclosing interface either
// as a *Named or an *Signature. Due to embedding, an interface may
// contain methods whose receiver type is a different interface.
// NOTE: Panic, if TypKind != Signature
func (fa *facade) Recv() *types.Var {
	return fa.signature().Recv()
}

// Results returns the results of signature s, or nil.
// NOTE: Panic, if TypKind != Signature
func (fa *facade) Results() *types.Tuple {
	return fa.signature().Results()
}

// Variadic reports whether the signature s is variadic.
// NOTE: Panic, if TypKind != Signature
func (fa *facade) Variadic() bool {
	return fa.signature().Variadic()
}

// Body returns function body.
// NOTE: Panic, if TypKind != Signature
func (fa *facade) Body() (string, error) {
	fa.signature()
	_, nodes, _ := fa.pkg.pathEnclosingInterval(fa.ident.Pos(), fa.ident.End())
	for _, node := range nodes {
		switch decl := node.(type) {
		case *ast.FuncDecl:
			return fa.pkg.prog.FormatNode(decl.Body)
		case *ast.FuncLit:
			return fa.pkg.prog.FormatNode(decl.Body)
		}
	}
	return "", errors.New("not found function body")
}

// CoverBody covers function body.
// NOTE: Panic, if TypKind != Signature
func (fa *facade) CoverBody(body string) error {
	fa.signature()
	switch decl := fa.Node().(type) {
	case *ast.FuncDecl:
		return fa.replaceFuncBody(fa.File(), decl.Body, body)
		// case *ast.FuncLit:
		// 	return errors.New("not support *ast.FuncLit")
	}
	return errors.New("not support")
}

func (fa *facade) replaceFuncBody(file *loader.File, node *ast.BlockStmt, newContent string) error {
	newContentBytes := []byte("package " + fa.pkg.Pkg.Name() + "\n" +
		strings.SplitN(fa.String(), "{", 2)[0] + "{" +
		strings.Replace(strings.TrimSpace(newContent), "\n", ";", -1) +
		"}")
	// TODO:
	// Possible file name conflicts
	// f, err := parser.ParseFile(fa.pkg.prog.fset, goutil.Md5(newContentBytes), newContentBytes, parser.ParseComments)
	f, err := parser.ParseFile(fa.pkg.prog.fset, file.Filename, newContentBytes, parser.ParseComments)
	if err != nil {
		return err
	}
	if len(f.Decls) != 1 {
		return errors.New("not support")
	}
	funcDecl, ok := f.Decls[0].(*ast.FuncDecl)
	if !ok || funcDecl.Body == nil {
		return errors.New("not support")
	}
	node.List = funcDecl.Body.List
	return nil
}

// func (fa *facade) replaceFile(file *loader.File, node ast.Node, newContent string) error {
// 	fileCode, err := fa.replaceCode(file, node, newContent)
// 	if err != nil {
// 		return nil
// 	}
// 	fset := token.NewFileSet()
// 	_, err = parser.ParseFile(fset, file.Filename, fileCode, parser.ParseComments)
// 	if err != nil {
// 		return nil
// 	}
// 	return nil
// }
// func (fa *facade) replaceCode(file *loader.File, node ast.Node, newContent string) ([]byte, error) {
// 	content, err := fa.pkg.prog.source(file.Filename)
// 	if err != nil {
// 		return nil, err
// 	}
// 	f := fa.pkg.prog.fset.File(node.Pos())
// 	if f == nil {
// 		return nil, errors.New("the node does not exist")
// 	}
// 	start := f.Offset(node.Pos())
// 	end := f.Offset(node.End())
// 	if start < 0 || (end >= 0 && start > end) {
// 		return content, nil
// 	}
// 	if end < 0 || end > len(content) {
// 		end = len(content)
// 	}
// 	if start > end {
// 		start = end
// 	}
// 	return bytes.Replace(content, content[start:end], goutil.StringToBytes(newContent), 1), nil
// }
