package parser

import (
	"github.com/sirupsen/logrus"
	"go/ast"
	"go/types"
	"reflect"
)

var ApiDeclParseFuncs ApiDeclParseFuncsType

func init() {
	ApiDeclParseFuncs = []ApiDeclParseFunc{
		CheckGinbuilderHandlerFuncDecl,
		CheckGinHandlerFuncDecl,
	}
}

// ApiDeclParseFunc 解析Api类型声明
type ApiDeclParseFunc func(
	decl ast.Decl,
	typesInfo *types.Info,
	parseRequestData bool,
) (apiItem *ApiItem, success bool, err error)

type ApiDeclParseFuncsType []ApiDeclParseFunc

func (m ApiDeclParseFuncsType) ParseApi(
	decl ast.Decl,
	typesInfo *types.Info,
	parseRequestData bool,
) (apiItem *ApiItem, success bool, err error) {
	for _, parseFunc := range m {
		apiItem, success, err = parseFunc(decl, typesInfo, parseRequestData)
		if err != nil {
			logrus.Errorf("parse api failed. func_type: %s, error: %s", reflect.TypeOf(parseFunc), err)
			return
		}

		if success {
			return
		}
	}

	return
}
