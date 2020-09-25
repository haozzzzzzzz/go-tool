package source

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/packages"
	"strings"
)

type ParsedType struct {
	TypesInfo *types.Info
	TypeName  *types.TypeName
	Doc       string
	Comment   string
}

type ParsedVal struct {
	TypeInfo *types.Info
	Type     types.Type
	Name     string
	Value    string
	Doc      string
	Comment  string
}

func Parse(
	dir string,
) (
	parsedTypes []*ParsedType,
	parsedVals []*ParsedVal,
	err error,
) {
	parsedTypes = make([]*ParsedType, 0)
	parsedVals = make([]*ParsedVal, 0)

	mode := packages.NeedImports |
		packages.NeedDeps |
		packages.NeedSyntax |
		packages.NeedTypesInfo |
		packages.NeedTypes |
		packages.NeedTypesSizes
	//mode := packages.NeedName |
	//	packages.NeedFiles |
	//	packages.NeedCompiledGoFiles |
	//	packages.NeedImports |
	//	packages.NeedDeps |
	//	packages.NeedExportsFile |
	//	packages.NeedTypes |
	//	packages.NeedSyntax |
	//	packages.NeedTypesInfo |
	//	packages.NeedTypesSizes
	pkgs, err := packages.Load(&packages.Config{
		Mode: mode,
		Dir:  dir,
	})
	if err != nil {
		logrus.Errorf("packages load failed. error: %s", err)
		return
	}

	for _, pkg := range pkgs {
		if pkg.TypesInfo == nil {
			logrus.Warnf("pkg has nil TypesInfo obj. name: %s", pkg.Name)
			continue
		}

		// 节点映射
		// 用于查找语法和注释
		specMapGenDecl := make(map[ast.Spec]*ast.GenDecl)
		for _, astFile := range pkg.Syntax {
			for _, decl := range astFile.Decls {
				genDecl, ok := decl.(*ast.GenDecl)
				if !ok {
					continue
				}

				for _, spec := range genDecl.Specs {
					specMapGenDecl[spec] = genDecl
				}
			}
		}

		// type definition
		for ident, def := range pkg.TypesInfo.Defs {
			switch defType := def.(type) {
			case *types.TypeName:
				spec, ok := ident.Obj.Decl.(*ast.TypeSpec)
				if !ok {
					logrus.Warnf("convert types.TypeName decl to *ast.TypeSpec not ok")
					break
				}

				typeDecl, ok := specMapGenDecl[spec]
				if !ok || typeDecl == nil {
					logrus.Warnf("type spec can not find type decl. name: %s", ident.Obj.Name)
					break
				}

				parsedTypes = append(parsedTypes, &ParsedType{
					TypesInfo: pkg.TypesInfo,
					TypeName:  defType,
					Doc:       strings.TrimSpace(typeDecl.Doc.Text()),
					Comment:   strings.TrimSpace(spec.Comment.Text()),
				})

			case *types.Const, *types.Var:
				spec, ok := ident.Obj.Decl.(*ast.ValueSpec)
				if !ok || spec == nil {
					break
				}

				varDecl, ok := specMapGenDecl[spec]
				if !ok || varDecl == nil {
					logrus.Warnf("value spec can not find decl. name: %s", ident.Obj.Name)
					break
				}

				val := ""
				if len(spec.Values) >= 1 {
					baseLit, ok := spec.Values[0].(*ast.BasicLit)
					if ok {
						val = baseLit.Value
					}
				}

				parsedVal := &ParsedVal{
					TypeInfo: pkg.TypesInfo,
					Type:     defType.Type(),
					Name:     defType.Name(),
					Value:    val,
					Doc:      strings.TrimSpace(varDecl.Doc.Text()), // TODO 目前还不能解析代码末尾的注释
					Comment:  strings.TrimSpace(spec.Comment.Text()),
				}

				parsedVals = append(parsedVals, parsedVal)

			default:
				break

			}

		}

	}

	return
}

// 类型解析器
var iTypeMap = map[types.Type]IType{}

func ParseType(
	info *types.Info,
	t types.Type,
) (iType IType) {
	var ok bool
	iType, ok = iTypeMap[t]
	if ok {
		return
	}

	iType = NewBasicType("Unsupported")

	switch t.(type) {
	case *types.Basic:
		iType = NewBasicType(t.(*types.Basic).Name())

	case *types.Pointer:
		iType = ParseType(info, t.(*types.Pointer).Elem())

	case *types.Named:
		tNamed := t.(*types.Named)
		iType = ParseType(info, tNamed.Underlying())

		// 如果是structType
		structType, ok := iType.(*StructType)
		if ok {
			structType.Name = tNamed.Obj().Name()
			iType = structType
		}

	case *types.Struct:
		structType := NewStructType()

		tStructType := t.(*types.Struct)

		typeAstExpr := FindStructAstExprFromInfoTypes(info, tStructType)
		if typeAstExpr == nil { // 找不到expr
			hasFieldsJsonTag := false
			numFields := tStructType.NumFields()
			for i := 0; i <= numFields; i++ {
				strTag := tStructType.Tag(i)
				mTagParts := parseStringTagParts(strTag)
				if len(mTagParts) == 0 {
					continue
				}

				for key, _ := range mTagParts {
					if key == "json" {
						hasFieldsJsonTag = true
						break
					}
				}

				if hasFieldsJsonTag {
					break
				}
			}

			if hasFieldsJsonTag { // 有导出的jsontag，但是找不到定义的
				logrus.Warnf("cannot found expr of type: %s", tStructType)
			}
		}

		numFields := tStructType.NumFields()
		for i := 0; i < numFields; i++ {
			field := NewField()

			tField := tStructType.Field(i)

			if !tField.Exported() && !tField.Embedded() { // 非嵌入类型，并且不是公开的，则跳过
				continue
			}

			if typeAstExpr != nil { // 找到声明
				astStructType, ok := typeAstExpr.(*ast.StructType)
				if !ok {
					logrus.Errorf("parse struct type failed. expr: %#v, type: %#v", typeAstExpr, tStructType)
					return
				}

				astField := astStructType.Fields.List[i]

				// 注释
				if astField.Doc != nil && len(astField.Doc.List) > 0 {
					for _, comment := range astField.Doc.List {
						if field.Description != "" {
							field.Description += "; "
						}

						field.Description += RemoveCommentStartEndToken(comment.Text)
					}
				}

				if astField.Comment != nil && len(astField.Comment.List) > 0 {
					for _, comment := range astField.Comment.List {
						if field.Description != "" {
							field.Description += "; "
						}
						field.Description += RemoveCommentStartEndToken(comment.Text)
					}
				}

			}

			// tags
			field.Tags = parseStringTagParts(tStructType.Tag(i))

			// field definition
			field.Name = tField.Name()
			fieldType := ParseType(info, tField.Type())
			field.TypeName = fieldType.TypeName()
			field.TypeSpec = fieldType
			field.Exported = tField.Exported() // 是否是公开的，内嵌类型为false
			field.Embedded = tField.Embedded() // 内嵌

			err := structType.AddFields(field)
			if nil != err {
				logrus.Warnf("parse struct type add field failed. error: %s.", err)
				return
			}

			// embedded
			if field.Embedded {
				structType.AddEmbedded(field)
			}

		}

		iType = structType

	case *types.Slice:
		arrType := NewArrayType()
		typeSlice := t.(*types.Slice)
		eltType := ParseType(info, typeSlice.Elem())
		arrType.EltSpec = eltType
		arrType.EltName = eltType.TypeName()
		arrType.Name = fmt.Sprintf("[]%s", eltType.TypeName())

		iType = arrType

	case *types.Array:
		arrType := NewArrayType()
		typeArr := t.(*types.Array)
		eltType := ParseType(info, typeArr.Elem())
		arrType.Len = typeArr.Len()
		arrType.EltSpec = eltType
		arrType.EltName = eltType.TypeName()
		if arrType.Len > 0 {
			arrType.Name = fmt.Sprintf("[%d]%s", arrType.Len, eltType.TypeName())
		} else {
			arrType.Name = fmt.Sprintf("[]%s", eltType.TypeName())
		}

		iType = arrType

	case *types.Map:
		mapType := NewMapType()
		tMap := t.(*types.Map)
		mapType.ValueSpec = ParseType(info, tMap.Elem())
		mapType.KeySpec = ParseType(info, tMap.Key())
		mapType.Name = fmt.Sprintf("map[%s]%s", mapType.KeySpec.TypeName(), mapType.ValueSpec.TypeName())

		iType = mapType

	case *types.Interface:
		iType = NewInterfaceType()

	default:
		logrus.Warnf("parse unsupported type %#v", t)

	}

	return
}

func parseStringTagParts(strTag string) (mParts map[string]string) {
	mParts = make(map[string]string, 0)
	tagValue := strings.Replace(strTag, "`", "", -1)
	strPairs := strings.Split(tagValue, " ")
	for _, pair := range strPairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		tagPair := strings.Split(pair, ":")
		mParts[tagPair[0]] = strings.Replace(tagPair[1], "\"", "", -1)
	}
	return

}

func convertExpr(expr ast.Expr) (newExpr ast.Expr) {
	switch expr.(type) {
	case *ast.StarExpr:
		newExpr = expr.(*ast.StarExpr).X

	case *ast.SelectorExpr:
		newExpr = expr.(*ast.SelectorExpr).Sel

	default:
		newExpr = expr

	}

	return
}

//var stop bool
// struct expr匹配类型
func FindStructAstExprFromInfoTypes(info *types.Info, t *types.Struct) (expr ast.Expr) {
	for tExpr, tType := range info.Types {
		var tStruct *types.Struct
		switch tType.Type.(type) {
		case *types.Struct:
			tStruct = tType.Type.(*types.Struct)
		}

		if tStruct == nil {
			continue
		}

		if t == tStruct {
			// 同一组astFiles生成的Types，内存中对象匹配成功
			expr = tExpr

			structExpr, ok := expr.(*ast.StructType)
			if ok {
				if t.NumFields() != len(structExpr.Fields.List) {
					fmt.Println(t)
					fmt.Println(tStruct)
					fmt.Println(tExpr)
				}
			}

		} else if t.String() == tStruct.String() {
			// 如果是不同的astFiles生成的Types，可能astFile1中没有这个类型信息，但是另外一组astFiles导入到info里，这是同一个类型，内存对象不一样，但是整体结构是一样的
			// 不是百分百准确
			expr = tExpr

		} else if tStruct.NumFields() == t.NumFields() {
			// 字段一样
			numFields := tStruct.NumFields()
			notMatch := false
			for i := 0; i < numFields; i++ {
				if tStruct.Tag(i) != t.Tag(i) || tStruct.Field(i).Name() != t.Field(i).Name() {
					notMatch = true
					break
				}
			}

			if notMatch == false { // 字段一样
				expr = tExpr
			}
		}

		_, isStructExpr := expr.(*ast.StructType)
		if isStructExpr {
			break
		}
	}

	return
}

func RemoveCommentStartEndToken(text string) (newText string) {
	newText = strings.Replace(text, "//", "", 1)
	newText = strings.Replace(newText, "/*", "", 1)
	newText = strings.Replace(newText, "*/", "", 1)
	newText = strings.TrimSpace(newText)
	return
}

// TypeIdent 类型定义或者类别名
type TypeIdent struct {
	TypeName *types.TypeName
	IType    IType
}

// 类型定义或者类别名
func ParseTypeName(
	info *types.Info,
	typeName *types.TypeName,
) (typeIdent *TypeIdent) {
	typeIdent = &TypeIdent{
		TypeName: typeName,
		IType:    ParseType(info, typeName.Type()), // 真实类型
	}

	return
}
