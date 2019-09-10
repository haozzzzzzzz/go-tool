package parser

import (
	"bufio"
	"errors"
	"github.com/haozzzzzzzz/go-rapid-development/utils/uerrors"
	"github.com/haozzzzzzzz/go-tool/api/com/mod"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"golang.org/x/tools/go/packages"
	"io"
	"path/filepath"
	"runtime/debug"
	"sort"

	"go/types"

	"os"

	"fmt"

	"strings"

	"github.com/go-playground/validator"
	"github.com/haozzzzzzzz/go-rapid-development/api/request"
	"github.com/haozzzzzzzz/go-rapid-development/utils/file"
	"github.com/sirupsen/logrus"
)

func (m *ApiParser) ScanApis(
	parseRequestData bool, // 如果parseRequestData会有点慢
	parseCommentText bool, // 是否从注释中提取api。`compile`不能从注释中生成routers
	usePackagesParse bool, // use Packages or Parser, gomod使用package，gopath使用parser
) (
	commonApiParams *ApiItemParams,
	apis []*ApiItem,
	err error,
) {
	apis = make([]*ApiItem, 0)

	// scan api dir
	apiDir, err := filepath.Abs(m.ApiDir)
	if nil != err {
		logrus.Warnf("get absolute file apiDir failed. \n%s.", err)
		return
	}

	commonParamsSlice, codeApis, err := ParseApis(apiDir, parseRequestData, parseCommentText, usePackagesParse)
	if nil != err {
		logrus.Errorf("parse apis from code failed. error: %s.", err)
		return
	}

	apis = append(apis, codeApis...)

	// sort api
	sortedApiUriKeys := make([]string, 0)
	mapApi := make(map[string]*ApiItem)
	for _, oneApi := range apis {
		if oneApi.RelativePaths == nil {
			continue
		}

		for _, relPath := range oneApi.RelativePaths {
			uriKey := m.apiUrlKey(relPath, oneApi.HttpMethod)
			sortedApiUriKeys = append(sortedApiUriKeys, uriKey)
			mapApi[uriKey] = oneApi
		}

	}

	sort.Strings(sortedApiUriKeys)

	sortedApis := make([]*ApiItem, 0)
	for _, key := range sortedApiUriKeys {
		sortedApis = append(sortedApis, mapApi[key])
	}
	apis = sortedApis

	// merge common params
	lenCommonParams := len(commonParamsSlice)
	if lenCommonParams > 1 {
		logrus.Warnf("found multi common params, merging them")
		commonApiParams, err = MergeApiItemParams(commonParamsSlice...)
		if nil != err {
			logrus.Errorf("merge apis item params failed. error: %s.", err)
			return
		}

	} else if lenCommonParams == 1 {
		commonApiParams = commonParamsSlice[0]

	}

	return
}

func ParseApis(
	apiDir string,
	parseRequestData bool, // 如果parseRequestData会有点慢
	parseCommentText bool, // 是否从注释中提取api。`compile`不能从注释中生成routers
	userPackageParse bool, // parseRequestData=true时，生效
) (
	commonParams []*ApiItemParams,
	apis []*ApiItem,
	err error,
) {
	apis = make([]*ApiItem, 0)
	logrus.Info("Scan api files...")
	defer func() {
		if err == nil {
			logrus.Info("Scan api files completed")
		}
	}()

	// api文件夹中所有的文件
	subApiDir := make([]string, 0)
	subApiDir, err = file.SearchFileNames(apiDir, func(fileInfo os.FileInfo) bool {
		if fileInfo.IsDir() {
			return true
		} else {
			return false
		}

	}, true)
	subApiDir = append(subApiDir, apiDir)

	// 服务源文件，只能一个pkg一个pkg地解析
	for _, subApiDir := range subApiDir {
		subCommonParams, subApis, errParse := ParsePkgApis(apiDir, subApiDir, userPackageParse, parseRequestData, parseCommentText)
		err = errParse
		if nil != err {
			logrus.Errorf("parse api file dir %q failed. error: %s.", subApiDir, err)
			return
		}

		apis = append(apis, subApis...)
		commonParams = append(commonParams, subCommonParams...)
	}

	return
}

func mergeTypesInfos(info *types.Info, infos ...*types.Info) {
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

func ParsePkgApis(
	apiRootDir string,
	apiPackageDir string,
	usePackagesParse bool, // use Packages or Parser
	parseRequestData bool,
	parseCommentText bool, // 是否从注释中提取api。`compile`不能从注释中生成routers
) (
	commonParams []*ApiItemParams,
	apis []*ApiItem,
	err error,
) {
	commonParams = make([]*ApiItemParams, 0)
	apis = make([]*ApiItem, 0)
	defer func() {
		if iRec := recover(); iRec != nil {
			logrus.Errorf("panic %s. api_dir: %s", iRec, apiPackageDir)
			debug.PrintStack()
		}
	}()

	hasGoFile, err := file.DirHasExtFile(apiPackageDir, ".go")
	if nil != err {
		logrus.Errorf("check dir has ext file failed. error: %s.", err)
		return
	}

	if !hasGoFile {
		return
	}

	goPaths := make([]string, 0)
	var goModName, goModDir string

	if !usePackagesParse {
		goPaths = strings.Split(os.Getenv("GOPATH"), ":")
		if len(goPaths) == 0 {
			err = uerrors.Newf("failed to find go paths")
			return
		}

	} else {
		_, goModName, goModDir = mod.FindGoMod(apiPackageDir)
		if goModName == "" || goModDir == "" {
			err = uerrors.Newf("failed to find go mod")
			return
		}
	}

	// 检索当前目录下所有的文件，不包含import的子文件
	astFiles := make([]*ast.File, 0)
	astFileNames := make(map[*ast.File]string, 0)

	// types
	typesInfo := &types.Info{
		Scopes:     make(map[ast.Node]*types.Scope),
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Implicits:  make(map[ast.Node]types.Object),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
		InitOrder:  make([]*types.Initializer, 0),
	}

	fileSet := token.NewFileSet()

	if !usePackagesParse { // not use mod
		parseMode := parser.AllErrors | parser.ParseComments
		astPkgMap, errParse := parser.ParseDir(fileSet, apiPackageDir, nil, parseMode)
		err = errParse
		if nil != err {
			logrus.Errorf("parser parse dir failed. error: %s.", err)
			return
		}
		_ = astPkgMap

		astFileMap := make(map[string]*ast.File)
		for pkgName, astPkg := range astPkgMap {
			_ = pkgName
			for fileName, pkgFile := range astPkg.Files {
				astFiles = append(astFiles, pkgFile)
				astFileNames[pkgFile] = fileName
				astFileMap[fileName] = pkgFile
			}
		}

		if parseRequestData {
			// parse types
			typesConf := types.Config{
				//Importer: importer.Default(),
			}

			typesConf.Importer = importer.For("source", nil)

			// cur package
			pkg, errCheck := typesConf.Check(apiPackageDir, fileSet, astFiles, typesInfo)
			err = errCheck
			if nil != err {
				logrus.Errorf("check types failed. error: %s.", err)
				return
			}
			_ = pkg

			// imported
			impPaths := make(map[string]bool)
			for fileName, pkgFile := range astFileMap {
				_ = fileName
				for _, fileImport := range pkgFile.Imports {
					impPath := strings.Replace(fileImport.Path.Value, "\"", "", -1)
					for _, goPath := range goPaths {
						absPath := fmt.Sprintf("%s/src/%s", goPath, impPath)
						if file.PathExists(absPath) {
							impPaths[absPath] = true
						}
					}
				}
			}

			for impPath, _ := range impPaths {
				tempFileSet := token.NewFileSet()
				tempAstFiles := make([]*ast.File, 0)

				tempPkgs, errParse := parser.ParseDir(tempFileSet, impPath, nil, parseMode)
				err = errParse
				if nil != err {
					logrus.Errorf("parser parse dir failed. error: %s.", err)
					return
				}

				for pkgName, pkg := range tempPkgs {
					_ = pkgName
					for _, pkgFile := range pkg.Files {
						tempAstFiles = append(tempAstFiles, pkgFile)
					}
				}

				// type check imported path
				_, err = typesConf.Check(impPath, tempFileSet, tempAstFiles, typesInfo)
				if nil != err {
					logrus.Errorf("check imported package failed. path: %s, error: %s.", impPath, err)
					return
				}

			}

		}

	} else { // use go mod
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
		pkgs, errLoad := packages.Load(&packages.Config{
			Mode: mode,
			Dir:  apiPackageDir,
			Fset: fileSet,
		})
		err = errLoad
		if nil != err {
			logrus.Errorf("go/packages load failed. error: %s.", err)
			return
		}

		// only one package
		pkgPath := make(map[string]bool)
		impPkgPaths := make(map[string]bool)
		for _, pkg := range pkgs {
			_, ok := pkgPath[pkg.PkgPath]
			if ok {
				continue
			}

			pkgPath[pkg.PkgPath] = true

			// 合并types.Info
			for _, astFile := range pkg.Syntax {
				astFiles = append(astFiles, astFile)
				astFileNames[astFile] = pkg.Fset.File(astFile.Pos()).Name()
			}

			// 包内的类型
			mergeTypesInfos(typesInfo, pkg.TypesInfo)

			if parseRequestData {
				for _, impPkg := range pkg.Imports { // 依赖
					_, ok := impPkgPaths[impPkg.PkgPath]
					if ok {
						continue
					}

					impPkgPaths[impPkg.PkgPath] = true

					// 包引用的包的类型，目前只解析一层
					mergeTypesInfos(typesInfo, impPkg.TypesInfo)
				}
			}
		}
	}

	// map api item def ident ast obj to types.Named
	mapApiObjTypes := make(map[*ast.Ident]*types.Named)
	for ident, def := range typesInfo.Defs {
		typesVar, ok := def.(*types.Var)
		if !ok {
			continue
		}

		typesNamed, ok := typesVar.Type().(*types.Named)
		if !ok {
			continue
		}

		typeName := typesNamed.Obj()
		typePkg := typeName.Pkg()
		if typePkg == nil || typePkg.Name() != "ginbuilder" || typeName.Name() != "HandleFunc" {
			continue
		}

		_, ok = mapApiObjTypes[ident]
		if ok {
			continue
		}

		// ginbuilder.HandlerFunc obj
		mapApiObjTypes[ident] = typesNamed
	}

	for _, astFile := range astFiles { // 遍历当前package的语法树
		fileName := astFileNames[astFile]
		logrus.Infof("Parsing %s", fileName)

		// package
		fileDir := filepath.Dir(fileName)

		var pkgRelAlias string
		packageName := astFile.Name.Name
		var pkgExportedPath string

		// 设置相对api文件夹的relative path
		pkgRelDir := strings.Replace(fileDir, apiRootDir, "", 1)
		if pkgRelDir != "" { // 子目录
			pkgRelDir = strings.Replace(pkgRelDir, "/", "", 1)
			pkgRelAlias = strings.Replace(pkgRelDir, "/", "_", -1)
		}

		// package exported path
		if !usePackagesParse { // 使用GOPATH
			foundInGoPath := false

			for _, subGoPath := range goPaths {
				if strings.Contains(fileDir, subGoPath) {
					foundInGoPath = true

					// 在此gopath下
					pkgExportedPath = strings.Replace(fileDir, subGoPath+"/src/", "", -1)
					break
				}
			}

			if !foundInGoPath {
				err = uerrors.Newf("failed to find file in go path. file: %s", fileName)
				return
			}

		} else { // 使用GOMOD
			pkgRelModPath := strings.ReplaceAll(fileDir, goModDir, "")
			pkgExportedPath = goModName + pkgRelModPath

		}

		if parseCommentText {
			// scan package comment apis
			for _, commentGroup := range astFile.Comments {
				for _, comment := range commentGroup.List {
					subCommonParams, commentApis, errParse := ParseApisFromPkgCommentText(
						fileName,
						fileDir,
						packageName,
						pkgExportedPath,
						pkgRelAlias,
						comment.Text,
					)
					err = errParse
					if nil != err {
						logrus.Errorf("parse apis from pkg comment text failed. error: %s.", err)
						return
					}

					apis = append(apis, commentApis...)

					commonParams = append(commonParams, subCommonParams...)
				}
			}
		}

		// search package api types
		for _, decl := range astFile.Decls /*objName, obj := range astFile.Scope.Objects*/ { // 遍历顶层所有变量，寻找HandleFunc
			genDel, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}

			if len(genDel.Specs) == 0 {
				return
			}

			valueSpec, ok := genDel.Specs[0].(*ast.ValueSpec)
			if !ok {
				continue
			}

			obj := valueSpec.Names[0] // variables with name
			objName := obj.Name
			selectorExpr, ok := valueSpec.Type.(*ast.SelectorExpr)
			if !ok {
				continue
			}

			xIdent, ok := selectorExpr.X.(*ast.Ident)
			if !ok {
				continue
			}

			selIdent := selectorExpr.Sel

			if xIdent.Name != "ginbuilder" && selIdent.Name != "HandleFunc" {
				continue
			}

			// ginbuild.HandlerFunc obj

			var summary, description string
			if genDel.Doc != nil {
				apiComment := genDel.Doc.Text()
				strBuf := bufio.NewReader(strings.NewReader(apiComment))
				bLine, _, errRead := strBuf.ReadLine()
				err = errRead
				if nil != err {
					logrus.Errorf("read api comment first line failed. error: %s.", err)
					return
				}

				summary = string(bLine)
				for {
					bDesc := make([]byte, 1000)
					num, errR := strBuf.Read(bDesc)
					err = errR
					if nil != err && err != io.EOF {
						logrus.Errorf("read description failed. error: %s.", err)
						return
					}

					if err == io.EOF {
						err = nil
						break
					}

					if num <= 0 {
						break
					}

					description += string(bDesc)
				}

				summary = strings.TrimSpace(summary)
				description = strings.TrimSpace(description)
			}

			apiType, ok := mapApiObjTypes[obj]
			if !ok {
				logrus.Warnf("failed to find types of api ast ident. %s", obj)
			}
			_ = apiType

			apiItem := &ApiItem{
				SourceFile:          fileName,
				ApiHandlerFunc:      objName,
				PackageName:         packageName,
				PackageExportedPath: pkgExportedPath,
				PackageRelAlias:     pkgRelAlias,
				PackageDir:          fileDir,
				RelativePaths:       make([]string, 0),
				Summary:             summary,
				Description:         description,
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
						funcBody := funcLit.Body
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

									switch ident.Name {
									case "headerData":
										apiItem.HeaderData = parseApiRequest(typesInfo, ident)
									case "pathData", "uriData":
										apiItem.UriData = parseApiRequest(typesInfo, ident)
									case "queryData":
										apiItem.QueryData = parseApiRequest(typesInfo, ident)
									case "postData":
										apiItem.PostData = parseApiRequest(typesInfo, ident)
									case "respData":
										apiItem.RespData = parseApiRequest(typesInfo, ident)
									}
								}

							case *ast.ReturnStmt:

							}

						}

					}

				}
			}

			err = validator.New().Struct(apiItem)
			if nil != err {
				logrus.Errorf("%#v\n invalid", apiItem)
				return
			}

			apis = append(apis, apiItem)
		}

	}

	return
}

func parseApiRequest(
	info *types.Info,
	astIdent *ast.Ident,
) (dataType *StructType) {
	identType := info.Defs[astIdent]
	typeVar, ok := identType.(*types.Var)
	if !ok {
		return
	}

	iType := parseType(info, typeVar.Type())
	if iType != nil {
		dataType, _ = iType.(*StructType)
	} else {
		logrus.Warnf("parse api request nil: %#v\n", typeVar)
	}

	return
}

func parseType(
	info *types.Info,
	t types.Type,
) (iType IType) {
	iType = NewBasicType("Unsupported")

	switch t.(type) {
	case *types.Basic:
		iType = NewBasicType(t.(*types.Basic).Name())

	case *types.Pointer:
		iType = parseType(info, t.(*types.Pointer).Elem())

	case *types.Named:
		tNamed := t.(*types.Named)
		iType = parseType(info, tNamed.Underlying())

		// 如果是structType
		structType, ok := iType.(*StructType)
		if ok {
			structType.Name = tNamed.Obj().Name()
			iType = structType
		}

	case *types.Struct: // 匿名
		structType := NewStructType()

		tStructType := t.(*types.Struct)

		typeAstExpr := FindStructAstExprFromInfoTypes(info, tStructType)
		if typeAstExpr == nil { // 不在当前源码内的定义会解析不到
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
			if !tField.Exported() {
				continue
			}

			if typeAstExpr != nil { // 找到声明
				astStructType, ok := typeAstExpr.(*ast.StructType)
				if !ok {
					logrus.Printf("parse struct type failed. expr: %#v, type: %#v", typeAstExpr, tStructType)
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

			// definition
			field.Name = tField.Name()
			fieldType := parseType(info, tField.Type())
			field.TypeName = fieldType.TypeName()
			field.TypeSpec = fieldType

			err := structType.AddFields(field)
			if nil != err {
				logrus.Warnf("parse struct type add field failed. error: %s.", err)
				return
			}

		}

		iType = structType

	case *types.Slice:
		arrType := NewArrayType()
		eltType := parseType(info, t.(*types.Slice).Elem())
		arrType.EltSpec = eltType
		arrType.EltName = eltType.TypeName()
		arrType.Name = fmt.Sprintf("[]%s", eltType.TypeName())

		iType = arrType

	case *types.Map:
		mapType := NewMapType()
		tMap := t.(*types.Map)
		mapType.ValueSpec = parseType(info, tMap.Elem())
		mapType.KeySpec = parseType(info, tMap.Key())
		mapType.Name = fmt.Sprintf("map[%s]%s", mapType.KeySpec.TypeName(), mapType.ValueSpec.TypeName())

		iType = mapType

	case *types.Interface:
		iType = NewInterfaceType()

	default:
		fmt.Printf("parse unsupported type %#v\n", t)

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
		tStruct, ok := tType.Type.(*types.Struct)
		if !ok {
			continue
		}

		if t == tStruct {
			// 同一组astFiles生成的Types，内存中对象匹配成功
			expr = tExpr

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
