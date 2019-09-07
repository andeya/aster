package aster

import (
	"go/ast"
	"go/types"
	"reflect"
	"strings"
	"unsafe"

	"github.com/henrylee2cn/goutil/tpack"
)

type invalidType struct {
	types.Type
	typeName string
}

func (t *invalidType) String() string {
	return t.typeName
}

func (fa *facade) checkParams() {
	fa.checkParamsOrResults("checkParams")
}

func (fa *facade) checkResults() {
	fa.checkParamsOrResults("checkResults")
}

func (fa *facade) checkParamsOrResults(op string) {
	defer func() { recover() }()
	sig := fa.signature()
	var fnType *ast.FuncType
	switch t := fa.Node().(type) {
	case *ast.FuncDecl:
		fnType = t.Type
	case *ast.FuncLit:
		fnType = t.Type
	default:
		return
	}
	var list []*ast.Field
	var tup *types.Tuple
	switch op {
	case "checkParams":
		list = fnType.Params.List
		tup = sig.Params()
	case "checkResults":
		list = fnType.Results.List
		tup = sig.Results()
	}
	for i := tup.Len() - 1; i >= 0; i-- {
		v := tup.At(i)
		if !strings.HasSuffix(v.Type().String(), "invalid type") {
			continue
		}
		typeName, err := fa.FormatNode(list[i].Type)
		if err != nil {
			continue
		}
		objVal := reflect.ValueOf(v).Elem().FieldByName("object")
		objTyp := objVal.FieldByName("typ")
		fakeType := &invalidType{
			Type:     v.Type(),
			typeName: typeName,
		}
		uptr := tpack.From(objTyp).Pointer()
		*(*types.Type)(unsafe.Pointer(uptr)) = fakeType
	}
}
