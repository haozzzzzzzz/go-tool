package parser

import (
	"github.com/haozzzzzzzz/go-rapid-development/utils/uerrors"
	"github.com/haozzzzzzzz/go-tool/common/source/mod"
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
				fileName := pkg.Fset.File(astFile.Pos()).Name()
				astFiles = append(astFiles, astFile)
				astFileNames[astFile] = fileName
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
			apiItem, success, errP := ApiDeclParseFuncs.ParseApi(decl, typesInfo, parseRequestData)
			err = errP
			if err != nil {
				logrus.Errorf("parse api declare failed. error: %s", err)
				return
			}

			if success && apiItem != nil {
				fileApis = append(fileApis, apiItem)
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
