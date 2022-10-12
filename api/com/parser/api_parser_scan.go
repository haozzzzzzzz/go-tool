package parser

import (
	"github.com/haozzzzzzzz/go-rapid-development/v2/utils/uerrors"
	"github.com/haozzzzzzzz/go-tool/common/source/mod"
	"go/ast"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/packages"
	"io/fs"
	"path/filepath"
	"runtime/debug"
	"sort"

	"strings"

	"github.com/haozzzzzzzz/go-rapid-development/v2/utils/file"
	"github.com/sirupsen/logrus"
)

func (m *ApiParser) ScanApis(
	parseRequestData bool, // 如果parseRequestData会有点慢
	parseCommentText bool, // 是否从注释中提取api。`compile`不能从注释中生成routers
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

	commonApiParamsMap, apis, err = ParseApis(apiDir, parseRequestData, parseCommentText)
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
		// required relative paths and http method
		if oneApi.RelativePaths == nil ||
			len(oneApi.RelativePaths) == 0 ||
			oneApi.HttpMethod == "" {
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
	subApiDirs := make([]string, 0)
	// TODO 需要优化。只需解析编译一次，不用每个目录都要解析编译
	err = filepath.WalkDir(apiDir, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			return nil
		}

		for _, skipDir := range SkipDirs {
			if d.Name() == skipDir {
				return fs.SkipDir
			}
		}

		subApiDirs = append(subApiDirs, path)
		return nil
	})
	if err != nil {
		logrus.Errorf("walk dir failed . dir: %s, err: %s", apiDir, err)
		return
	}

	// 服务源文件，只能一个pkg一个pkg地解析
	for _, subApiDir := range subApiDirs {
		logrus.Infof("parse package %s", subApiDir)
		subCommonParamsList, subApis, errParse := ParsePkgApis(apiDir, subApiDir, parseRequestData, parseCommentText)
		err = errParse
		if nil != err {
			logrus.Errorf("parse api file dir %q failed. error: %s.", subApiDir, err)
			return
		}

		apis = append(apis, subApis...)

		for _, subCommonParams := range subCommonParamsList {
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

var SkipDirs = []string{
	".git",
	".idea",
}

// SkipFileNameSuffixes skip files with suffix
var SkipFileNameSuffixes = []string{
	"test.go",
	"/api/routers.go",
}

func ParsePkgApis(
	apiRootDir string,
	apiPackageDir string,
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

	_, goModName, goModDir := mod.FindGoMod(apiPackageDir)
	if goModName == "" || goModDir == "" {
		err = uerrors.Newf("failed to find go mod")
		return
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

	mode := packages.NeedName |
		packages.NeedFiles |
		packages.NeedCompiledGoFiles |
		packages.NeedImports |
		packages.NeedDeps |
		//packages.NeedExportFile |
		packages.NeedTypes |
		packages.NeedSyntax |
		packages.NeedTypesInfo |
		packages.NeedTypesSizes |
		packages.NeedModule |
		packages.NeedEmbedFiles

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

	//logrus.Infof("load packages : %s, pkgs: %+v", apiPackageDir, pkgs)

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

		// 解析请求结构
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

	for _, astFile := range astFiles { // 遍历当前package的语法树
		fileApis := make([]*ApiItem, 0)
		fileName := astFileNames[astFile]

		// skip *test.go and routers.go
		needSkip := false
		for _, suffix := range SkipFileNameSuffixes {
			if strings.HasSuffix(fileName, suffix) {
				needSkip = true
				break
			}
		}

		if needSkip {
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

		// 使用GOMOD
		pkgRelModPath := strings.ReplaceAll(fileDir, goModDir, "")
		pkgExportedPath = goModName + pkgRelModPath

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
