package containers

import (
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"github.com/unknown321/datfpk/fox2/datatypes/fox"
	"github.com/unknown321/datfpk/util"
	"io"
	"log/slog"

	"github.com/unknown321/hashing"
)

type FoxStringMapEntry struct {
	Key       uint64
	KeyString string
	Value     fox.DataType
}

type FoxStringMap struct {
	Data     []FoxStringMapEntry
	DataType fox.FDataType
	next     int
}

func (f *FoxStringMap) Read(reader io.ReadSeeker) error {
	for i := 0; i < len(f.Data); i++ {
		switch f.DataType {
		case fox.FInt8:
			f.Data[i].Value = &fox.Int8{}
		case fox.FString:
			f.Data[i].Value = &fox.String{}
		case fox.FEntityHandle:
			f.Data[i].Value = &fox.EntityHandle{}
		case fox.FEntityPtr:
			f.Data[i].Value = &fox.EntityPtr{}
		case fox.FFilePtr:
			f.Data[i].Value = &fox.FilePtr{}
		case fox.FPath:
			f.Data[i].Value = &fox.Path{}
		case fox.FUInt8:
			f.Data[i].Value = &fox.UInt8{}
		case fox.FInt16:
			f.Data[i].Value = &fox.Int16{}
		case fox.FUInt16:
			f.Data[i].Value = &fox.UInt16{}
		case fox.FUInt32:
			f.Data[i].Value = &fox.UInt32{}
		case fox.FInt32:
			f.Data[i].Value = &fox.Int32{}
		case fox.FUInt64:
			f.Data[i].Value = &fox.UInt64{}
		case fox.FInt64:
			f.Data[i].Value = &fox.Int64{}
		case fox.FDouble:
			f.Data[i].Value = &fox.Double{}
		case fox.FBool:
			f.Data[i].Value = &fox.Bool{}
		case fox.FVector3:
			f.Data[i].Value = &fox.Vector3{}
		case fox.FVector4:
			f.Data[i].Value = &fox.Vector4{}
		case fox.FMatrix3:
			f.Data[i].Value = &fox.Matrix3{}
		case fox.FMatrix4:
			f.Data[i].Value = &fox.Matrix4{}
		case fox.FColor:
			f.Data[i].Value = &fox.Color{}
		case fox.FQuat:
			f.Data[i].Value = &fox.Quat{}
		case fox.FEntityLink:
			f.Data[i].Value = &fox.EntityLink{}
		case fox.FFloat:
			f.Data[i].Value = &fox.Float{}
		case fox.FWideVector3:
			f.Data[i].Value = &fox.WideVector3{}
		default:
			panic(fmt.Sprintf("stringMap, not implemented: %s", f.DataType))
		}

		if err := binary.Read(reader, binary.LittleEndian, &f.Data[i].Key); err != nil {
			return fmt.Errorf("stringmap key: %w", err)
		}

		if err := f.Data[i].Value.Read(reader); err != nil {
			return fmt.Errorf("stringmap value: %w", err)
		}

		if _, err := util.AlignRead(reader, 16); err != nil {
			return fmt.Errorf("stringmap align: %w", err)
		}

		//slog.Info("stringMap", "i", i, "key", fmt.Sprintf("%x", f.Data[i].Key), "value", fmt.Sprintf("%+v", f.Data[i].Value), "dataType", f.DataType)
	}

	return nil
}

func (f *FoxStringMap) Write(writer io.WriteSeeker) error {
	var err error
	for _, v := range f.Data {
		if v.Value == nil {
			return fmt.Errorf("stringMap value is nil, key \"%s\"", v.KeyString)
		}
		v.Key = hashing.StrCode64([]byte(v.KeyString))
		if err = binary.Write(writer, binary.LittleEndian, v.Key); err != nil {
			return fmt.Errorf("stringMap key: %w", err)
		}
		if err = v.Value.Write(writer); err != nil {
			return fmt.Errorf("stringMap value: %w", err)
		}
		if _, err = util.AlignWrite(writer, 16); err != nil {
			return fmt.Errorf("stringMap align: %w", err)
		}
	}
	return nil
}

func (f *FoxStringMap) Next() func() *fox.DataType {
	i := 0
	return func() *fox.DataType {
		return &f.Data[i].Value
	}
}

func NewFoxStringMap(dataType fox.FDataType, count int) *FoxStringMap {
	f := &FoxStringMap{
		Data:     make([]FoxStringMapEntry, count),
		DataType: dataType,
	}

	for i := range f.Data {
		switch f.DataType {
		case fox.FInt8:
			f.Data[i].Value = &fox.Int8{}
		case fox.FString:
			f.Data[i].Value = &fox.String{}
		case fox.FEntityHandle:
			f.Data[i].Value = &fox.EntityHandle{}
		case fox.FEntityPtr:
			f.Data[i].Value = &fox.EntityPtr{}
		case fox.FFilePtr:
			f.Data[i].Value = &fox.FilePtr{}
		case fox.FPath:
			f.Data[i].Value = &fox.Path{}
		case fox.FUInt8:
			f.Data[i].Value = &fox.UInt8{}
		case fox.FInt16:
			f.Data[i].Value = &fox.Int16{}
		case fox.FUInt16:
			f.Data[i].Value = &fox.UInt16{}
		case fox.FUInt32:
			f.Data[i].Value = &fox.UInt32{}
		case fox.FInt32:
			f.Data[i].Value = &fox.Int32{}
		case fox.FUInt64:
			f.Data[i].Value = &fox.UInt64{}
		case fox.FInt64:
			f.Data[i].Value = &fox.Int64{}
		case fox.FDouble:
			f.Data[i].Value = &fox.Double{}
		case fox.FBool:
			f.Data[i].Value = &fox.Bool{}
		case fox.FVector3:
			f.Data[i].Value = &fox.Vector3{}
		case fox.FVector4:
			f.Data[i].Value = &fox.Vector4{}
		case fox.FMatrix3:
			f.Data[i].Value = &fox.Matrix3{}
		case fox.FMatrix4:
			f.Data[i].Value = &fox.Matrix4{}
		case fox.FColor:
			f.Data[i].Value = &fox.Color{}
		case fox.FQuat:
			f.Data[i].Value = &fox.Quat{}
		case fox.FEntityLink:
			f.Data[i].Value = &fox.EntityLink{}
		case fox.FFloat:
			f.Data[i].Value = &fox.Float{}
		case fox.FWideVector3:
			f.Data[i].Value = &fox.WideVector3{}
		default:
			panic(fmt.Sprintf("stringMap, not implemented: %s", f.DataType))
		}
	}

	return f
}

type fsmEntry struct {
	Key      string        `xml:"key,attr"`
	Data     fox.DataType  `xml:"data"`
	DataType fox.FDataType `xml:"-"`
}

func (f *fsmEntry) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	for _, a := range start.Attr {
		//slog.Info("attrib", "v", fmt.Sprintf("%s = %s", a.Name.Local, a.Value))
		switch a.Name.Local {
		case "key":
			f.Key = a.Value
		default:
			slog.Warn("ignored stringMap entry attribute %s, value", a.Name.Local, a.Value)
		}
	}

	for {
		t, err := d.Token()
		if err != nil {
			panic(err)
		}

		switch tt := t.(type) {
		case xml.StartElement:
			//slog.Info("xml", "tag", fmt.Sprintf("<%s> %+v\n", tt.Name.Local, tt.Attr))
			if f.Data, err = DecodeFoxData(d, &tt, f.DataType); err != nil {
				return err
			}
		case xml.EndElement:
			//slog.Info("xml", "tag", fmt.Sprintf("</%s>", tt.Name.Local))
			if tt == start.End() {
				return nil
			}
		}
	}

	return nil
}

func DecodeFoxData(decoder *xml.Decoder, start *xml.StartElement, dType fox.FDataType) (fox.DataType, error) {
	var data fox.DataType
	switch dType {
	case fox.FInt8:
		data = &fox.Int8{}
	case fox.FString:
		data = &fox.String{}
	case fox.FEntityHandle:
		data = &fox.EntityHandle{}
	case fox.FEntityPtr:
		data = &fox.EntityPtr{}
	case fox.FFilePtr:
		data = &fox.FilePtr{}
	case fox.FPath:
		data = &fox.Path{}
	case fox.FUInt8:
		data = &fox.UInt8{}
	case fox.FInt16:
		data = &fox.Int16{}
	case fox.FUInt16:
		data = &fox.UInt16{}
	case fox.FUInt32:
		data = &fox.UInt32{}
	case fox.FInt32:
		data = &fox.Int32{}
	case fox.FUInt64:
		data = &fox.UInt64{}
	case fox.FInt64:
		data = &fox.Int64{}
	case fox.FDouble:
		data = &fox.Double{}
	case fox.FBool:
		data = &fox.Bool{}
	case fox.FVector3:
		data = &fox.Vector3{}
	case fox.FVector4:
		data = &fox.Vector4{}
	case fox.FMatrix3:
		data = &fox.Matrix3{}
	case fox.FMatrix4:
		data = &fox.Matrix4{}
	case fox.FColor:
		data = &fox.Color{}
	case fox.FQuat:
		data = &fox.Quat{}
	case fox.FEntityLink:
		data = &fox.EntityLink{}
	case fox.FFloat:
		data = &fox.Float{}
	case fox.FWideVector3:
		data = &fox.WideVector3{}
	default:
		panic(fmt.Sprintf("stringMap, not implemented: %s", dType))
	}

	if err := decoder.DecodeElement(data, start); err != nil {
		return nil, err
	}

	return data, nil
}

func (f *FoxStringMap) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	elements := make([]fsmEntry, len(f.Data))
	for i, v := range f.Data {
		elements[i].Key = v.KeyString
		elements[i].Data = v.Value
	}

	return e.EncodeElement(elements, start)
}

func (f *FoxStringMap) DecodeNext(d *xml.Decoder, start *xml.StartElement) error {
	ee := &fsmEntry{DataType: f.DataType}
	if err := d.DecodeElement(ee, start); err != nil {
		return err
	}
	f.Data[f.next].Value = ee.Data
	f.Data[f.next].KeyString = ee.Key
	//slog.Info("mm", "data", fmt.Sprintf("%+v", ee.Data), "key", f.Data[f.next].KeyString)
	f.next++
	if f.next == len(f.Data) {
		f.next = 0
	}

	return nil
}

func (f *FoxStringMap) GetStrings() []string {
	res := []string{}
	for _, v := range f.Data {
		res = append(res, v.KeyString)
		if v.Value.String() != nil {
			res = append(res, v.Value.String()...)
		}
	}
	return res
}
