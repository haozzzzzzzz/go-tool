package source

import (
	"go/types"
)

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
