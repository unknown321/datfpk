package containers

import "fmt"

type FoxContainerType byte

//go:generate stringer -type=FoxContainerType

const (
	StaticArray FoxContainerType = iota
	DynamicArray
	StringMap
	List

	ContainerTypeUnknown
)

func ContainerTypeFromString(s string) (FoxContainerType, error) {
	for i, v := range _FoxContainerType_index {
		if i+1 == len(_FoxContainerType_index) {
			return ContainerTypeUnknown, fmt.Errorf("unknown type %s", s)
		}

		ss := _FoxContainerType_name[v:_FoxContainerType_index[i+1]]
		if ss == s {
			return FoxContainerType(i), nil
		}
	}

	return ContainerTypeUnknown, fmt.Errorf("unknown type %s", s)
}
