// Code generated by "stringer -type=FoxContainerType"; DO NOT EDIT.

package containers

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[StaticArray-0]
	_ = x[DynamicArray-1]
	_ = x[StringMap-2]
	_ = x[List-3]
}

const _FoxContainerType_name = "StaticArrayDynamicArrayStringMapList"

var _FoxContainerType_index = [...]uint8{0, 11, 23, 32, 36}

func (i FoxContainerType) String() string {
	if i >= FoxContainerType(len(_FoxContainerType_index)-1) {
		return "FoxContainerType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _FoxContainerType_name[_FoxContainerType_index[i]:_FoxContainerType_index[i+1]]
}
