package source

import (
	"fmt"
	"github.com/haozzzzzzzz/go-tool/common/source/mod"
	"github.com/sirupsen/logrus"
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/packages"
	"path/filepath"
	"strings"
)

type ParsedType struct {
	TypeName *types.TypeName
	Doc      string
	Comment  string
}

type ParsedVal struct {
	Type    types.Type
	Name    string
	Value   string
	Doc     string
	Comment string
}

func Parse(
	dir string,
) (
	mergedTypesInfo *types.Info,
	parsedTypes []*ParsedType,
	parsedVals []*ParsedVal,
	err error,
) {
	parsedTypes = make([]*ParsedType, 0)
	parsedVals = make([]*ParsedVal, 0)

	// 查询项目go mod
	_, goModName, goModDir := mod.FindGoMod(dir)
	_ = goModName

	mode := packages.NeedName |
		packages.NeedFiles |
		packages.NeedCompiledGoFiles |
		packages.NeedImports |
		packages.NeedDeps |
		packages.NeedExportsFile |
		packages.NeedTypes |
		packages.NeedSyntax |
		packages.NeedTypesInfo |
		packages.NeedTypesSizes

	scanDir := dir
	if goModDir != "" {
		scanDir = goModDir
	}

	pkgs, err := packages.Load(&packages.Config{
		Mode: mode,
		Dir:  scanDir,
	})

	if err != nil {
		logrus.Errorf("packages load failed. error: %s", err)
		return
	}

	pkgMap := make(map[string]*packages.Package)
	for _, pkg := range pkgs {
		pkgMap[pkg.ID] = pkg
		importPkgMap := importPackages(pkg)
		for iPkgId, iPkg := range importPkgMap {
			pkgMap[iPkgId] = iPkg
		}
	}

	parsePkgs := make([]*packages.Package, 0)
	mergedTypesInfo = &types.Info{
		Scopes:     make(map[ast.Node]*types.Scope),
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Implicits:  make(map[ast.Node]types.Object),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
		InitOrder:  make([]*types.Initializer, 0),
	}

	for pkgName, pkg := range pkgMap {
		MergeTypesInfos(mergedTypesInfo, pkg.TypesInfo)

		// 只解析指定目录下的类型
		if len(pkg.GoFiles) == 0 {
			continue
		}

		_ = pkgName
		goFileDir := filepath.Dir(pkg.GoFiles[0])
		if strings.Contains(goFileDir, dir) {
			parsePkgs = append(parsePkgs, pkg)
		}
	}

	for _, pkg := range parsePkgs {
		fmt.Printf("Parsing package %s\n", pkg.ID)

		pkgTypes, pkgVals, errParse := ParsePackage(pkg)
		err = errParse
		if err != nil {
			logrus.Errorf("parse package failed. pkg: %s, error: %s", pkg.ID, err)
			return
		}

		parsedTypes = append(parsedTypes, pkgTypes...)
		parsedVals = append(parsedVals, pkgVals...)

	}

	return
}

func importPackages(
	pkg *packages.Package,
) (pkgMap map[string]*packages.Package) {
	pkgMap = make(map[string]*packages.Package)
	if len(pkg.Imports) == 0 {
		return
	}

	for _, importPkg := range pkg.Imports {
		pkgMap[importPkg.ID] = importPkg
		nextLevelPkgMap := importPackages(importPkg)
		for _, nPkg := range nextLevelPkgMap {
			pkgMap[nPkg.ID] = nPkg
		}
	}

	return
}

func MergeTypesInfos(info *types.Info, infos ...*types.Info) {
	for _, tempInfo := range infos {
		for tKey, tVal := range tempInfo.Types {
			info.Types[tKey] = tVal
		}

		for defKey, defVal := range tempInfo.Defs {
			info.Defs[defKey] = defVal
		}

		for useKey, useVal := range tempInfo.Uses {
			info.Uses[useKey] = useVal
		}

		for implKey, implVal := range tempInfo.Implicits {
			info.Implicits[implKey] = implVal
		}

		for selKey, selVal := range tempInfo.Selections {
			info.Selections[selKey] = selVal
		}

		for scopeKey, scopeVal := range tempInfo.Scopes {
			info.Scopes[scopeKey] = scopeVal
		}

		// do not need to merge InitOrder
	}
	return
}

// 从一个包里解析类型
func ParsePackage(
	pkg *packages.Package,
) (
	parsedTypes []*ParsedType,
	parsedVals []*ParsedVal,
	err error,
) {
	parsedTypes = make([]*ParsedType, 0)
	parsedVals = make([]*ParsedVal, 0)

	astFiles := pkg.Syntax
	typesInfo := pkg.TypesInfo

	// 节点映射
	// 用于查找语法和注释
	specMapGenDecl := make(map[ast.Spec]*ast.GenDecl)
	for _, astFile := range astFiles {
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
	for ident, def := range typesInfo.Defs {
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
				TypeName: defType,
				Doc:      strings.TrimSpace(typeDecl.Doc.Text()),
				Comment:  strings.TrimSpace(spec.Comment.Text()),
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
				Type:    defType.Type(),
				Name:    defType.Name(),
				Value:   val,
				Doc:     strings.TrimSpace(varDecl.Doc.Text()), // TODO 目前还不能解析代码末尾的注释
				Comment: strings.TrimSpace(spec.Comment.Text()),
			}

			parsedVals = append(parsedVals, parsedVal)

		default:
			break

		}

	}

	return
}

var ITypeMap = map[types.Type]IType{} // 防止重复解析；防止内部循环引用时，一直无法停止

// 类型解析器
func ParseType(
	info *types.Info,
	t types.Type,
) (iType IType) {
	iType, ok := ITypeMap[t] // 可能是类型一样，但是名字不一样的类
	if ok {                  // 如果类型解析过了，则退出
		return
	}

	switch t.(type) {
	case *types.Basic:
		iType = NewBasicType(t.(*types.Basic).Name())
		ITypeMap[t] = iType

	case *types.Pointer:
		iType = ParseType(info, t.(*types.Pointer).Elem())
		ITypeMap[t] = iType

	case *types.Named:
		tNamed := t.(*types.Named)
		iType = ParseType(info, tNamed.Underlying())
		ITypeMap[t] = iType // 记录底层类

		// 如果是structType
		// 可能内部ITypeMap使用同一个对象，则复制一个出来，避免多个使用type A B重新定义的类名字被改写
		structType, ok := iType.(*StructType)
		if ok {
			cloneStruct := structType.Copy()
			cloneStruct.Name = tNamed.Obj().Name()
			iType = cloneStruct // 返回复制类
		}

	case *types.Struct:
		structType := NewStructType()
		iType = structType
		ITypeMap[t] = iType

		tStructType := t.(*types.Struct)

		// 从抽象树中查找结构体的声明
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
				logrus.Warnf("cannot find expr of type: %s", tStructType)
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

	case *types.Slice:
		arrType := NewArrayType()
		iType = arrType
		ITypeMap[t] = iType

		typeSlice := t.(*types.Slice)

		// 内部可能发生循环引用，所以将递归放到最后, 保证第一遍递归的时候上层类信息尽量全
		eltType := ParseType(info, typeSlice.Elem())
		arrType.EltSpec = eltType
		arrType.EltName = eltType.TypeName()
		arrType.Name = fmt.Sprintf("[]%s", eltType.TypeName())

	case *types.Array:
		arrType := NewArrayType()
		iType = arrType
		ITypeMap[t] = iType

		typeArr := t.(*types.Array)
		arrType.Len = typeArr.Len()

		// 内部可能发生循环引用，所以将递归放到最后, 保证第一遍递归的时候上层类信息尽量全
		eltType := ParseType(info, typeArr.Elem())
		arrType.EltSpec = eltType
		arrType.EltName = eltType.TypeName()
		if arrType.Len > 0 {
			arrType.Name = fmt.Sprintf("[%d]%s", arrType.Len, eltType.TypeName())
		} else {
			arrType.Name = fmt.Sprintf("[]%s", eltType.TypeName())
		}

	case *types.Map:
		mapType := NewMapType()
		iType = mapType
		ITypeMap[t] = iType

		tMap := t.(*types.Map)

		// 内部可能发生循环引用，所以将递归放到最后, 保证第一遍递归的时候上层类信息尽量全
		mapType.ValueSpec = ParseType(info, tMap.Elem())
		mapType.KeySpec = ParseType(info, tMap.Key())
		mapType.Name = fmt.Sprintf("map[%s]%s", mapType.KeySpec.TypeName(), mapType.ValueSpec.TypeName())

	case *types.Interface:
		iType = NewInterfaceType()
		ITypeMap[t] = iType

	default:
		iType = NewBasicType("Unsupported")
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

		separateIndex := strings.Index(pair, ":")
		if separateIndex < 0 || separateIndex == len(pair)-1 {
			continue
		}

		key := pair[:separateIndex]
		value := pair[separateIndex+1:]

		mParts[key] = strings.Replace(value, "\"", "", -1)
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
