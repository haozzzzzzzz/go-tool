package parser

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator"
	request "github.com/haozzzzzzzz/go-rapid-development/api/request"
	"github.com/sirupsen/logrus"
	"go/ast"
	"go/types"
	"strings"
)

// 解析ginbuilder
func CheckGinbuilderHandlerFuncDecl(
	decl ast.Decl,
	typesInfo *types.Info,
	parseRequestData bool,
) (
	apiItem *ApiItem,
	success bool,
	err error,
) {
	apiItem = NewApiItem()
	genDel, ok := decl.(*ast.GenDecl)
	if !ok {
		return
	}

	if len(genDel.Specs) == 0 {
		return
	}

	valueSpec, ok := genDel.Specs[0].(*ast.ValueSpec)
	if !ok {
		return
	}

	obj := valueSpec.Names[0] // variables with name
	selectorExpr, ok := valueSpec.Type.(*ast.SelectorExpr)
	if !ok {
		return
	}

	xIdent, ok := selectorExpr.X.(*ast.Ident)
	if !ok {
		return
	}

	selIdent := selectorExpr.Sel

	if !(xIdent.Name == "ginbuilder" && selIdent.Name == "HandleFunc") {
		return
	}

	err = ParseGinbuilderHandleFuncApi(
		apiItem,
		genDel,
		valueSpec,
		obj,
		typesInfo,
		parseRequestData,
	)
	if nil != err {
		logrus.Errorf("parse ginbuild handle func api failed. error: %s.", err)
		return
	}

	success = true
	return
}

// 解析ginbuilder.HandleFunc
// https://github.com/haozzzzzzzz/go-rapid-development/tree/master/web/ginbuilder
func ParseGinbuilderHandleFuncApi(
	apiItem *ApiItem,
	genDel *ast.GenDecl,
	valueSpec *ast.ValueSpec,
	obj *ast.Ident, // variables with name
	typesInfo *types.Info,
	parseRequestData bool,
) (err error) {
	// ginbuilder.HandlerFunc obj
	apiItem.ApiHandlerFunc = obj.Name
	apiItem.ApiHandlerFuncType = ApiHandlerFuncTypeGinbuilderHandleFunc

	if genDel.Doc != nil {
		apiComment := genDel.Doc.Text()
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

	for _, value := range valueSpec.Values { // 遍历属性
		compositeLit, ok := value.(*ast.CompositeLit)
		if !ok {
			continue
		}

		// compositeLit.Elts
		for _, elt := range compositeLit.Elts {
			keyValueExpr, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}

			keyIdent, ok := keyValueExpr.Key.(*ast.Ident)
			if !ok {
				continue
			}

			switch keyIdent.Name {
			case "HttpMethod":
				valueLit, ok := keyValueExpr.Value.(*ast.BasicLit)
				if !ok {
					break
				}

				value := strings.Replace(valueLit.Value, "\"", "", -1)
				value = strings.ToUpper(value)
				switch value {
				case request.METHOD_GET,
					request.METHOD_POST,
					request.METHOD_PUT,
					request.METHOD_PATCH,
					request.METHOD_HEAD,
					request.METHOD_OPTIONS,
					request.METHOD_DELETE,
					request.METHOD_CONNECT,
					request.METHOD_TRACE,
					request.METHOD_ANY:

				default:
					err = errors.New(fmt.Sprintf("unsupported http method : %s", value))
					logrus.Errorf("mapping unsupported api failed. %s.", err)
					return
				}

				apiItem.HttpMethod = value

			case "RelativePath": // 废弃
				valueLit, ok := keyValueExpr.Value.(*ast.BasicLit)
				if !ok {
					break
				}

				value := strings.Replace(valueLit.Value, "\"", "", -1)
				apiItem.RelativePaths = append(apiItem.RelativePaths, value)

			case "RelativePaths":
				compLit, ok := keyValueExpr.Value.(*ast.CompositeLit)
				if !ok {
					break
				}

				for _, elt := range compLit.Elts {
					basicLit, ok := elt.(*ast.BasicLit)
					if !ok {
						continue
					}

					value := strings.Replace(basicLit.Value, "\"", "", -1)
					apiItem.RelativePaths = append(apiItem.RelativePaths, value)
				}

			case "Handle":
				funcLit, ok := keyValueExpr.Value.(*ast.FuncLit)
				if !ok {
					break
				}

				// if parse request data
				if parseRequestData == false {
					break
				}

				// parse request data
				parseApiFuncBody(apiItem, funcLit.Body, typesInfo)

			}

		}
	}

	err = validator.New().Struct(apiItem)
	if nil != err {
		logrus.Errorf("%#v\n invalid", apiItem)
		return
	}

	return
}
