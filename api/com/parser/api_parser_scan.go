package parser

import (
	"errors"
	"github.com/haozzzzzzzz/go-rapid-development/utils/uerrors"
	"github.com/haozzzzzzzz/go-tool/api/com/mod"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/packages"
	"os"
	"path/filepath"
	"runtime/debug"
	"sort"

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
	commonApiParamsMap map[string]*ApiItemParams, // dir -> common params
	apis []*ApiItem,
	err error,
) {
	// scan api dir
	apiDir, err := filepath.Abs(m.ApiDir)
	if nil != err {
		logrus.Warnf("get absolute file apiDir failed. \n%s.", err)
		return
	}

	commonApiParamsMap, apis, err = ParseApis(apiDir, parseRequestData, parseCommentText, usePackagesParse)
	if nil != err {
		logrus.Errorf("parse apis from code failed. error: %s.", err)
		return
	}

	// merge common params
	if parseRequestData && len(commonApiParamsMap) > 0 {
		for _, api := range apis {
			pkgDir := api.ApiFile.PackageDir
			matchedCommonParams := make([]*ApiItemParams, 0)
			for dir, commonParams := range commonApiParamsMap {
				if strings.Contains(pkgDir, dir) {
					matchedCommonParams = append(matchedCommonParams, commonParams)
				}
			}

			if len(matchedCommonParams) == 0 {
				continue
			}

			for _, commonParams := range matchedCommonParams {
				err = api.ApiItemParams.MergeApiItemParams(commonParams)
				if nil != err {
					logrus.Errorf("api merge common params failed. %s", err)
					return
				}
			}
		}
	}

	// sort api
	sortedApiUriKeys := make([]string, 0)
	uriKeyExistsMap := make(map[string]*ApiItem)
	mapApi := make(map[string]*ApiItem)
	for _, oneApi := range apis {
		if oneApi.RelativePaths == nil || len(oneApi.RelativePaths) == 0 || oneApi.HttpMethod == "" { // required relative paths and http method
			logrus.Warnf("api need to declare relative paths and http method. func name: %s", oneApi.ApiHandlerFunc)
			continue
		}

		// check url path if exists
		for _, path := range oneApi.RelativePaths {
			pathUriKey := m.apiUrlKey(path, oneApi.HttpMethod)
			existUrlKeyItem, ok := uriKeyExistsMap[pathUriKey]
			if ok {
				err = uerrors.Newf("duplicated url handler. 1: %s, 2: %s", existUrlKeyItem.ApiHandlerFuncPath(), oneApi.ApiHandlerFuncPath())
				return

			} else {
				uriKeyExistsMap[pathUriKey] = oneApi
			}
		}

		firstRelPath := oneApi.RelativePaths[0] // use first relative paths to sort
		uriKey := m.apiUrlKey(firstRelPath, oneApi.HttpMethod)
		sortedApiUriKeys = append(sortedApiUriKeys, uriKey)
		mapApi[uriKey] = oneApi
	}

	sort.Strings(sortedApiUriKeys)

	sortedApis := make([]*ApiItem, 0)
	for _, key := range sortedApiUriKeys {
		sortedApis = append(sortedApis, mapApi[key])
	}

	apis = sortedApis

	return
}

func ParseApis(
	apiDir string,
	parseRequestData bool, // 如果parseRequestData会有点慢
	parseCommentText bool, // 是否从注释中提取api。`compile`不能从注释中生成routers
	userPackageParse bool, // parseRequestData=true时，生效
) (
	commonParamsMap map[string]*ApiItemParams, // file dir -> api item params
	apis []*ApiItem,
	err error,
) {
	commonParamsMap = make(map[string]*ApiItemParams)
	apis = make([]*ApiItem, 0)
	logrus.Info("Scan api files ...")
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

		subCommonParamses, subApis, errParse := ParsePkgApis(apiDir, subApiDir, userPackageParse, parseRequestData, parseCommentText)
		err = errParse
		if nil != err {
			logrus.Errorf("parse api file dir %q failed. error: %s.", subApiDir, err)
			return
		}

		apis = append(apis, subApis...)

		for _, subCommonParams := range subCommonParamses {
			_, ok := commonParamsMap[subApiDir]
			if !ok {
				commonParamsMap[subApiDir] = subCommonParams
			} else {
				err = commonParamsMap[subApiDir].MergeApiItemParams(subCommonParams)
				if nil != err {
					logrus.Errorf("merge api item common params failed. %#v", subCommonParams)
				}
			}
		}
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
			for fileName, pkgFile := range astFileMap { // only parse first level of imported
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

			// parse ast files and types
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

	for _, astFile := range astFiles { // 遍历当前package的语法树
		fileApis := make([]*ApiItem, 0)
		fileName := astFileNames[astFile]

		// skip *_test.go and routers.go
		if strings.HasSuffix(fileName, "_test.go") || strings.HasSuffix(fileName, "/api/routers.go") {
			logrus.Infof("Skip parsing %s", fileName)
			continue
		}

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

		apiFile := NewApiFile()
		apiFile.SourceFile = fileName
		apiFile.PackageName = packageName
		apiFile.PackageExportedPath = pkgExportedPath
		apiFile.PackageRelAlias = pkgRelAlias
		apiFile.PackageDir = fileName

		// search package api types
		for _, decl := range astFile.Decls { // 遍历顶层所有变量，寻找HandleFunc
			apiItem := &ApiItem{
				RelativePaths: make([]string, 0),
			}

			switch decl.(type) {
			case *ast.GenDecl: // var xx = yy
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
				selectorExpr, ok := valueSpec.Type.(*ast.SelectorExpr)
				if !ok {
					continue
				}

				xIdent, ok := selectorExpr.X.(*ast.Ident)
				if !ok {
					continue
				}

				selIdent := selectorExpr.Sel

				if !(xIdent.Name == "ginbuilder" && selIdent.Name == "HandleFunc") {
					continue
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

				fileApis = append(fileApis, apiItem)

			case *ast.FuncDecl: // type HandlerFunc func(*Context)
				funcDecl := decl.(*ast.FuncDecl)

				// 判断是否是gin.HandlerFunc
				paramsList := funcDecl.Type.Params.List
				if len(paramsList) != 1 || funcDecl.Type.Results != nil {
					continue
				}

				paramField := paramsList[0]
				startExpr, ok := paramField.Type.(*ast.StarExpr)
				if !ok {
					continue
				}

				selectExpr, ok := startExpr.X.(*ast.SelectorExpr)
				if !ok || selectExpr.X == nil || selectExpr.Sel == nil {
					continue
				}

				pkgIdent := selectExpr.X.(*ast.Ident)
				objIdent := selectExpr.Sel

				if !(pkgIdent.Name == "gin" && objIdent.Name == "Context") {
					continue
				}

				err = ParseGinHandlerFuncApi(apiItem, funcDecl, typesInfo, parseRequestData)
				if nil != err {
					logrus.Errorf("parse gin HandlerFunc failed. error: %s.", err)
					return
				}

				fileApis = append(fileApis, apiItem)

			default:
				continue

			}

		}

		// parse comment text
		if parseCommentText {
			// parse file package comment
			if astFile.Doc != nil {
				pkgComment := astFile.Doc.Text()
				pkgTags, errParse := ParseCommentTags(pkgComment)
				err = errParse
				if nil != err {
					logrus.Errorf("parse pkg comment tags failed. error: %s.", err)
					return
				}

				if pkgTags != nil {
					apiFile.MergeInfoFromCommentTags(pkgTags)
				}

			}

			// scan file comment apis
			for _, commentGroup := range astFile.Comments {
				for _, comment := range commentGroup.List {
					subCommonParams, commentApis, errParse := ParseApisFromPkgCommentText(
						comment.Text,
					)
					err = errParse
					if nil != err {
						logrus.Errorf("parse apis from pkg comment text failed. error: %s.", err)
						return
					}

					fileApis = append(fileApis, commentApis...)

					commonParams = append(commonParams, subCommonParams...)
				}
			}

		}

		for _, api := range fileApis {
			api.ApiFile = apiFile
			api.MergeInfoFromApiInfo()
			apis = append(apis, api)
		}

	}

	return
}

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
	parseApiFuncBody(apiItem, funcDecl.Body, typesInfo)

	return

}

func parseApiFuncBody(
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
					iType := parseType(typesInfo, typeVar.Type())
					if iType == nil {
						continue
					}

					structType, ok := iType.(*StructType)
					if ok {
						apiItem.HeaderData = structType
					} else {
						logrus.Warnf("header data is not struct type")
					}

				case "pathData", "uriData":
					iType := parseType(typesInfo, typeVar.Type())
					if iType == nil {
						continue
					}

					structType, ok := iType.(*StructType)
					if ok {
						apiItem.UriData = structType
					} else {
						logrus.Warnf("uri data is not struct type")
					}

				case "queryData":
					iType := parseType(typesInfo, typeVar.Type())
					if iType == nil {
						continue
					}

					structType, ok := iType.(*StructType)
					if ok {
						apiItem.QueryData = structType
					} else {
						logrus.Warnf("query data is not struct type")
					}

				case "postData":
					iType := parseType(typesInfo, typeVar.Type())
					if iType == nil {
						continue
					}

					apiItem.PostData = iType

				case "respData":
					iType := parseType(typesInfo, typeVar.Type())
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
			fieldType := parseType(info, tField.Type())
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
		eltType := parseType(info, typeSlice.Elem())
		arrType.EltSpec = eltType
		arrType.EltName = eltType.TypeName()
		arrType.Name = fmt.Sprintf("[]%s", eltType.TypeName())

		iType = arrType

	case *types.Array:
		arrType := NewArrayType()
		typeArr := t.(*types.Array)
		eltType := parseType(info, typeArr.Elem())
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
		mapType.ValueSpec = parseType(info, tMap.Elem())
		mapType.KeySpec = parseType(info, tMap.Key())
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
