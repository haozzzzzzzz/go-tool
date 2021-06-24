package parser

import (
	"github.com/haozzzzzzzz/go-tool/common/source"
	"github.com/sirupsen/logrus"
	"go/ast"
	"go/types"
)

/**
parse gin style api
https://github.com/gin-gonic/gin
*/

func CheckGinHandlerFuncDecl(
	decl ast.Decl,
	typesInfo *types.Info,
	parseRequestData bool,
) (
	apiItem *ApiItem,
	success bool,
	err error,
) {
	apiItem = NewApiItem()
	funcDecl, ok := decl.(*ast.FuncDecl)
	if !ok {
		return
	}

	// 判断是否是gin.HandlerFunc
	paramsList := funcDecl.Type.Params.List
	if len(paramsList) != 1 || funcDecl.Type.Results != nil {
		return
	}

	paramField := paramsList[0]
	startExpr, ok := paramField.Type.(*ast.StarExpr)
	if !ok {
		return
	}

	selectExpr, ok := startExpr.X.(*ast.SelectorExpr)
	if !ok || selectExpr.X == nil || selectExpr.Sel == nil {
		return
	}

	pkgIdent := selectExpr.X.(*ast.Ident)
	objIdent := selectExpr.Sel

	if !(pkgIdent.Name == "gin" && objIdent.Name == "Context") {
		return
	}

	err = ParseGinHandlerFuncApi(apiItem, funcDecl, typesInfo, parseRequestData)
	if nil != err {
		logrus.Errorf("parse gin HandlerFunc failed. error: %s.", err)
		return
	}

	success = true
	return
}

// 解析gin handler
// https://github.com/gin-gonic/gin
func ParseGinHandlerFuncApi(
	apiItem *ApiItem,
	funcDecl *ast.FuncDecl,
	typesInfo *types.Info,
	parseRequestData bool,
) (err error) {
	apiItem.ApiHandlerFunc = funcDecl.Name.Name
	apiItem.ApiHandlerFuncType = ApiHandlerFuncTypeGinHandlerFunc

	// 读取注释
	if funcDecl.Doc != nil {
		apiComment := funcDecl.Doc.Text()
		commentTags, errParse := ParseCommentTags(apiComment)
		err = errParse
		if nil != err {
			logrus.Errorf("parse api comment tags failed. error: %s.", err)
			return
		}

		if commentTags != nil {
			apiItem.MergeInfoFomCommentTags(commentTags)
		}
	}

	if !parseRequestData {
		return
	}

	// parse request data
	ParseApiFuncBody(apiItem, funcDecl.Body, typesInfo)

	return

}

// 解析handler里的请求结构和返回结构
func ParseApiFuncBody(
	apiItem *ApiItem,
	funcBody *ast.BlockStmt,
	typesInfo *types.Info,
) {
	for _, funcStmt := range funcBody.List {
		switch funcStmt.(type) {
		case *ast.AssignStmt:
			assignStmt := funcStmt.(*ast.AssignStmt)
			lhs := assignStmt.Lhs
			rhs := assignStmt.Rhs

			_ = lhs
			_ = rhs

			for _, expr := range lhs {
				ident, ok := expr.(*ast.Ident)
				if !ok {
					continue
				}

				identType := typesInfo.Defs[ident]
				typeVar, ok := identType.(*types.Var)
				if !ok {
					continue
				}

				switch ident.Name {
				case "headerData":
					iType := source.ParseType(typesInfo, typeVar.Type())
					if iType == nil {
						continue
					}

					structType, ok := iType.(*source.StructType)
					if ok {
						apiItem.HeaderData = structType
					} else {
						logrus.Warnf("header data is not struct type")
					}

				case "pathData", "uriData":
					iType := source.ParseType(typesInfo, typeVar.Type())
					if iType == nil {
						continue
					}

					structType, ok := iType.(*source.StructType)
					if ok {
						apiItem.UriData = structType
					} else {
						logrus.Warnf("uri data is not struct type")
					}

				case "queryData":
					iType := source.ParseType(typesInfo, typeVar.Type())
					if iType == nil {
						continue
					}

					structType, ok := iType.(*source.StructType)
					if ok {
						apiItem.QueryData = structType
					} else {
						logrus.Warnf("query data is not struct type")
					}

				case "postData":
					iType := source.ParseType(typesInfo, typeVar.Type())
					if iType == nil {
						continue
					}

					apiItem.PostData = iType

				case "respData":
					iType := source.ParseType(typesInfo, typeVar.Type())
					if iType == nil {
						continue
					}

					apiItem.RespData = iType

				}
			}

		case *ast.ReturnStmt:

		}

	}

	return
}
