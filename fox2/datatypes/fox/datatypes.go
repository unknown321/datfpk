package fox

import (
	"fmt"
	"io"
)

type DataType interface {
	Read(reader io.Reader) error
	Write(writer io.Writer) error

	String() []string
	//HashString() string // used in stringMap when String() returns ""

	Resolve(map[uint64]string)
}

//go:generate stringer -type=FDataType

type FDataType byte

const (
	FInt8 FDataType = iota
	FUInt8
	FInt16
	FUInt16
	FInt32
	FUInt32
	FInt64
	FUInt64
	FFloat
	FDouble
	FBool
	FString
	FPath
	FEntityPtr
	FVector3
	FVector4
	FQuat
	FMatrix3
	FMatrix4
	FColor
	FFilePtr
	FEntityHandle
	FEntityLink
	FPropertyInfo // not implemented in FoxTool
	FWideVector3

	FFail
)

func DataTypeToString(t FDataType) string {
	return t.String()[1:]
}

func DataTypeFromString(s string) (FDataType, error) {
	for i, v := range _FDataType_index {
		if i+1 == len(_FDataType_index) {
			return FFail, fmt.Errorf("unknown type %s", s)
		}

		ss := _FDataType_name[v:_FDataType_index[i+1]]
		if ss[1:] == s {
			return FDataType(i), nil
		}
	}

	return FFail, fmt.Errorf("unknown type %s", s)
}
