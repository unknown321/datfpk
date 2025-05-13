package fox2

import (
	"datfpk/fox2/containers"
	"datfpk/fox2/datatypes/fox"
	"encoding/xml"
	"fmt"
	"io"
)

func ReadContainer(cont *IFoxContainer, reader io.ReadSeeker, dataType fox.FDataType, containerType containers.FoxContainerType, valueCount int16) error {
	var err error

	*cont, err = CreateTypedContainer(dataType, containerType, int(valueCount))
	if err != nil {
		return fmt.Errorf("create typed container: %w", err)
	}

	if err = (*cont).Read(reader); err != nil {
		return fmt.Errorf("container read: %w", err)
	}

	return nil
}

type IFoxContainer interface {
	Read(reader io.ReadSeeker) error
	Write(writer io.WriteSeeker) error
	Next() func() *fox.DataType
	DecodeNext(d *xml.Decoder, start *xml.StartElement) error
	GetStrings() []string
}

func CreateTypedContainer(dataType fox.FDataType, containerType containers.FoxContainerType, count int) (IFoxContainer, error) {
	var c IFoxContainer

	switch containerType {
	case containers.StaticArray:
		c = containers.NewFoxStaticArray(dataType, count)
	case containers.StringMap:
		c = containers.NewFoxStringMap(dataType, count)
	case containers.List:
		c = containers.NewFoxList(dataType, count)
	case containers.DynamicArray:
		c = containers.NewFoxDynamicArray(dataType, count)
	default:
		panic(fmt.Sprintf("container not implemented: %s", containerType))
	}

	return c, nil
}
