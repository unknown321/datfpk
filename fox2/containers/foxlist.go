package containers

import (
	"encoding/xml"
	"fmt"
	"github.com/unknown321/datfpk/fox2/datatypes/fox"
	"io"
)

type FoxList struct {
	Data []fox.DataType
	next int
}

func NewFoxList(dataType fox.FDataType, count int) *FoxList {
	data := make([]fox.DataType, count)
	for i := 0; i < count; i++ {
		switch dataType {
		case fox.FInt8:
			data[i] = &fox.Int8{}
		case fox.FString:
			data[i] = &fox.String{}
		case fox.FEntityHandle:
			data[i] = &fox.EntityHandle{}
		case fox.FEntityPtr:
			data[i] = &fox.EntityPtr{}
		case fox.FFilePtr:
			data[i] = &fox.FilePtr{}
		case fox.FPath:
			data[i] = &fox.Path{}
		case fox.FUInt8:
			data[i] = &fox.UInt8{}
		case fox.FInt16:
			data[i] = &fox.Int16{}
		case fox.FUInt16:
			data[i] = &fox.UInt16{}
		case fox.FUInt32:
			data[i] = &fox.UInt32{}
		case fox.FInt32:
			data[i] = &fox.Int32{}
		case fox.FUInt64:
			data[i] = &fox.UInt64{}
		case fox.FInt64:
			data[i] = &fox.Int64{}
		case fox.FDouble:
			data[i] = &fox.Double{}
		case fox.FBool:
			data[i] = &fox.Bool{}
		case fox.FVector3:
			data[i] = &fox.Vector3{}
		case fox.FVector4:
			data[i] = &fox.Vector4{}
		case fox.FMatrix3:
			data[i] = &fox.Matrix3{}
		case fox.FMatrix4:
			data[i] = &fox.Matrix4{}
		case fox.FColor:
			data[i] = &fox.Color{}
		case fox.FQuat:
			data[i] = &fox.Quat{}
		case fox.FEntityLink:
			data[i] = &fox.EntityLink{}
		case fox.FFloat:
			data[i] = &fox.Float{}
		case fox.FWideVector3:
			data[i] = &fox.WideVector3{}
		default:
			panic(fmt.Sprintf("not implemented %s", dataType))
		}
	}

	return &FoxList{Data: data}
}

func (f *FoxList) Next() func() *fox.DataType {
	i := 0
	return func() *fox.DataType {
		return &f.Data[i]
	}
}

func (f *FoxList) DecodeNext(d *xml.Decoder, start *xml.StartElement) error {
	err := d.DecodeElement(f.Data[f.next], start)
	f.next++
	if f.next == len(f.Data) {
		f.next = 0
	}
	return err
}

func (f *FoxList) Read(reader io.ReadSeeker) error {
	var err error
	for i := range f.Data {
		if err = f.Data[i].Read(reader); err != nil {
			return fmt.Errorf("foxList read: %w", err)
		}
	}

	return nil
}

func (f *FoxList) Write(writer io.WriteSeeker) error {
	for _, v := range f.Data {
		if err := v.Write(writer); err != nil {
			return fmt.Errorf("foxList value: %w", err)
		}
	}

	return nil
}

func (f *FoxList) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(f.Data, start)
}

func (f *FoxList) GetStrings() []string {
	res := make([]string, 0)
	for _, p := range f.Data {
		if p.String() != nil {
			res = append(res, p.String()...)
		}
	}

	return res
}
