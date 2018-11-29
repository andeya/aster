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

func init() {
}

// func pickFuncType(n ast.Node) bool {
// 	*ast.FuncDecl
// }

// // collectStructs collects and maps structType nodes to their positions
// func collectStructs(node ast.Node) map[token.Pos]*structType {
// 	structs := make(map[token.Pos]*structType, 0)
// 	collectStructs := func(n ast.Node) bool {
// 		var t ast.Expr
// 		var structName string

// 		switch x := n.(type) {
// 		case *ast.TypeSpec:
// 			if x.Type == nil {
// 				return true

// 			}

// 			structName = x.Name.Name
// 			t = x.Type
// 		case *ast.CompositeLit:
// 			t = x.Type
// 		case *ast.ValueSpec:
// 			structName = x.Names[0].Name
// 			t = x.Type
// 		}

// 		x, ok := t.(*ast.StructType)
// 		if !ok {
// 			return true
// 		}

// 		structs[x.Pos()] = &structType{
// 			name: structName,
// 			node: x,
// 		}
// 		return true
// 	}
// 	ast.Inspect(node, collectStructs)
// 	return structs
// }

// // collectStructs collects and maps structType nodes to their positions
// func (t *tplInfo) collectStructs() {
// 	collectStructs := func(n ast.Node) bool {
// 		decl, ok := n.(ast.Decl)
// 		if !ok {
// 			return true
// 		}
// 		genDecl, ok := decl.(*ast.GenDecl)
// 		if !ok {
// 			return true
// 		}
// 		var groupDoc string
// 		if len(genDecl.Specs) == 1 {
// 			groupDoc = genDecl.Doc.Text()
// 		}
// 		for _, spec := range genDecl.Specs {
// 			var e ast.Expr
// 			var structName string
// 			var doc = groupDoc

// 			switch x := spec.(type) {
// 			case *ast.TypeSpec:
// 				if x.Type == nil {
// 					continue
// 				}
// 				structName = x.Name.Name
// 				e = x.Type
// 				if s := x.Doc.Text(); s != "" {
// 					doc = x.Doc.Text()
// 				}
// 			}

// 			x, ok := e.(*ast.StructType)
// 			if !ok {
// 				continue
// 			}

// 			if len(x.Fields.List) == 0 {
// 				switch structName {
// 				case MYSQL_MODEL, MONGO_MODEL:
// 				default:
// 					if goutil.IsExportedName(structName) {
// 						a := &aliasType{
// 							doc:  addSlash(doc),
// 							name: structName,
// 							text: fmt.Sprintf("%s = codec.PbEmpty", structName),
// 						}
// 						a.rawTypeName = a.text[strings.LastIndex(strings.TrimSpace(strings.Split(a.text, "//")[0]), " ")+1:]
// 						if a.doc == "" {
// 							a.doc = fmt.Sprintf("// %s alias of type %s\n", a.name, a.rawTypeName)
// 						}
// 						t.aliasTypes = append(t.aliasTypes, a)
// 					}
// 					continue
// 				}
// 			}

// 			t.realStructTypes = append(
// 				t.realStructTypes,
// 				structType{
// 					name: structName,
// 					doc:  addSlash(doc),
// 					node: x,
// 				}.init(t),
// 			)
// 		}
// 		return true
// 	}
// 	ast.Inspect(t.astFile, collectStructs)
// 	t.sortStructs()
// }
