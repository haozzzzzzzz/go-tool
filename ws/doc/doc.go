package doc

import (
	"github.com/haozzzzzzzz/go-rapid-development/v2/utils/uerrors"
	"github.com/haozzzzzzzz/go-tool/ws/parse"
	"github.com/sirupsen/logrus"
)

// 将ws写入doc
func WriteDoc(
	wsTypes *parse.WsTypesOutput,
	format string,
	filepath string,
) (err error) {
	writer, ok := FileFormatWriterMap[format]
	if !ok || writer == nil {
		err = uerrors.Newf("unsupported format %s error: %s", format, err)
		return
	}

	err = writer(wsTypes, format, filepath)
	if err != nil {
		logrus.Errorf("writer write doc failed. error: %s", err)
		return
	}

	return
}

type FileFormatWriter func(wsTypes *parse.WsTypesOutput, format string, filepath string) (err error)

var FileFormatWriterMap = make(map[string]FileFormatWriter)

func BindFileFormatWriter(formats []string, writer FileFormatWriter) {
	for _, format := range formats {
		FileFormatWriterMap[format] = writer
	}
}

func GetSupportedFileFormat() (formats []string) {
	formats = make([]string, 0)
	for format, _ := range FileFormatWriterMap {
		formats = append(formats, format)
	}
	return
}
