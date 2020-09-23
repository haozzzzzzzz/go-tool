package source

// 类型解析器
type TypesParser struct {
	rootDir string
}

func NewTypesParser(
	rootDir string, // 根目录
) (parser *TypesParser) {
	parser = &TypesParser{
		rootDir: rootDir,
	}
	return
}

type ParseTypeFilter func() (match bool)

func (m *TypesParser) Parse(filter ParseTypeFilter) (err error) {
	// TODO
	return
}
