// Code generated by "stringer -type=FDataType"; DO NOT EDIT.

package fox

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[FInt8-0]
	_ = x[FUInt8-1]
	_ = x[FInt16-2]
	_ = x[FUInt16-3]
	_ = x[FInt32-4]
	_ = x[FUInt32-5]
	_ = x[FInt64-6]
	_ = x[FUInt64-7]
	_ = x[FFloat-8]
	_ = x[FDouble-9]
	_ = x[FBool-10]
	_ = x[FString-11]
	_ = x[FPath-12]
	_ = x[FEntityPtr-13]
	_ = x[FVector3-14]
	_ = x[FVector4-15]
	_ = x[FQuat-16]
	_ = x[FMatrix3-17]
	_ = x[FMatrix4-18]
	_ = x[FColor-19]
	_ = x[FFilePtr-20]
	_ = x[FEntityHandle-21]
	_ = x[FEntityLink-22]
	_ = x[FPropertyInfo-23]
	_ = x[FWideVector3-24]
	_ = x[FFail-25]
}

const _FDataType_name = "FInt8FUInt8FInt16FUInt16FInt32FUInt32FInt64FUInt64FFloatFDoubleFBoolFStringFPathFEntityPtrFVector3FVector4FQuatFMatrix3FMatrix4FColorFFilePtrFEntityHandleFEntityLinkFPropertyInfoFWideVector3FFail"

var _FDataType_index = [...]uint8{0, 5, 11, 17, 24, 30, 37, 43, 50, 56, 63, 68, 75, 80, 90, 98, 106, 111, 119, 127, 133, 141, 154, 165, 178, 190, 195}

func (i FDataType) String() string {
	if i >= FDataType(len(_FDataType_index)-1) {
		return "FDataType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _FDataType_name[_FDataType_index[i]:_FDataType_index[i+1]]
}
