package fox2

import (
	"datfpk/fox2/containers"
	"datfpk/util"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/unknown321/hashing"
)

type EntityHeader struct {
	HeaderSize           int16
	Unknown1             int16
	_                    int16
	Magic1               uint32
	Address              uint32
	_                    uint32
	Unknown2             int32
	Unknown5             int32
	Version              int16
	ClassNameHash        uint64
	StaticPropertyCount  uint16
	DynamicPropertyCount uint16
	Offset               int32
	StaticDataSize       int32
	DataSize             int32
}

var EntityHeaderSize int16 = 64
var EntityHeaderMagic uint32 = 0x746e65 // ent

func (eh *EntityHeader) Read(reader io.Reader) error {
	return binary.Read(reader, binary.LittleEndian, eh)
}

func (eh *EntityHeader) Write(writer io.Writer) error {
	return binary.Write(writer, binary.LittleEndian, eh)
}

type Entity struct {
	Header            EntityHeader
	StaticProperties  []Property
	DynamicProperties []Property
	ClassNameString   string
}

func (e *Entity) Resolve(dict map[uint64]string) {
	var ok bool
	e.ClassNameString, ok = dict[e.Header.ClassNameHash]
	if !ok {
		e.ClassNameString = fmt.Sprintf("0x%X", e.Header.ClassNameHash)
	}

	e.ResolveProps(dict)
}

type entityXml struct {
	Class             string     `xml:"class,attr"`
	ClassVersion      int16      `xml:"classVersion,attr"`
	Addr              string     `xml:"addr,attr"`
	Unknown1          int16      `xml:"unknown1,omitempty,attr"`
	Unknown2          int32      `xml:"unknown2,omitempty,attr"`
	Unknown5          int32      `xml:"unknown5,omitempty,attr"`
	StaticProperties  []Property `xml:"staticProperties>property"`
	DynamicProperties []Property `xml:"dynamicProperties>property"`
}

func (e *Entity) MarshalXML(encoder *xml.Encoder, start xml.StartElement) error {
	//slog.Info("marshal", "class", e.ClassNameString)
	xx := entityXml{
		Class:             e.ClassNameString,
		ClassVersion:      e.Header.Version,
		Addr:              fmt.Sprintf("0x%X", e.Header.Address),
		Unknown1:          e.Header.Unknown1,
		Unknown2:          e.Header.Unknown2,
		Unknown5:          e.Header.Unknown5,
		StaticProperties:  e.StaticProperties,
		DynamicProperties: e.DynamicProperties,
	}

	return encoder.EncodeElement(xx, start)
}

func (e *Entity) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var err error
	xx := entityXml{}
	if err = d.DecodeElement(&xx, &start); err != nil {
		return err
	}

	e.ClassNameString = xx.Class
	e.Header.Version = xx.ClassVersion
	addr, err := strconv.ParseInt(strings.TrimPrefix(xx.Addr, "0x"), 16, 32)
	if err != nil {
		return err
	}
	e.Header.Address = uint32(addr)

	e.Header.Unknown1 = xx.Unknown1
	e.Header.Unknown2 = xx.Unknown2
	e.Header.Unknown5 = xx.Unknown5
	e.StaticProperties = xx.StaticProperties
	e.DynamicProperties = xx.DynamicProperties

	return nil
}

func (e *Entity) ResolveProps(dict map[uint64]string) {
	var ok bool
	for i := range e.StaticProperties {
		e.StaticProperties[i].NameValue, ok = dict[e.StaticProperties[i].Header.NameHash]
		if !ok {
			e.StaticProperties[i].NameValue = fmt.Sprintf("0x%X", e.StaticProperties[i].Header.NameHash)
		}

		switch e.StaticProperties[i].Header.ContainerType {
		case containers.StringMap:
			sm := (e.StaticProperties[i].Value).(*containers.FoxStringMap)
			for n := range sm.Data {
				sm.Data[n].KeyString, ok = dict[sm.Data[n].Key]
				if !ok {
					sm.Data[n].KeyString = fmt.Sprintf("0x%X", sm.Data[n].Key)
				}

				//slog.Info("resolving", "value", fmt.Sprintf("%+v", sm.Data[n].Value))
				sm.Data[n].Value.Resolve(dict)
			}
		case containers.StaticArray:
			sm := (e.StaticProperties[i].Value).(*containers.FoxStaticArray)
			for e := range sm.Data {
				sm.Data[e].Resolve(dict)
			}
		case containers.DynamicArray:
			sm := (e.StaticProperties[i].Value).(*containers.FoxDynamicArray)
			for e := range sm.Data {
				sm.Data[e].Resolve(dict)
			}
		case containers.List:
			sm := (e.StaticProperties[i].Value).(*containers.FoxList)
			for e := range sm.Data {
				sm.Data[e].Resolve(dict)
			}
		}
	}

	for i := range e.DynamicProperties {
		e.DynamicProperties[i].NameValue, ok = dict[e.DynamicProperties[i].Header.NameHash]
		if !ok {
			e.DynamicProperties[i].NameValue = fmt.Sprintf("0x%X", e.DynamicProperties[i].Header.NameHash)
		}
	}
}

func (e *Entity) Read(reader io.ReadSeeker) error {
	var err error
	if err = e.Header.Read(reader); err != nil {
		return fmt.Errorf("header: %w", err)
	}

	if _, err = util.AlignRead(reader, 16); err != nil {
		return err
	}

	//o, _ := reader.Seek(0, io.SeekCurrent)
	//slog.Info("entity read", "offset", o, "v", fmt.Sprintf("%+v", e.Header))

	//slog.Info("entity prop",
	//	"static count", e.Header.StaticPropertyCount,
	//	"size", e.Header.StaticDataSize,
	//	"dynamic count", e.Header.DynamicPropertyCount,
	//	"total size", e.Header.DataSize)

	for i := 0; i < int(e.Header.StaticPropertyCount); i++ {
		//o, _ = reader.Seek(0, io.SeekCurrent)
		//slog.Info("read prop at", "v", o)
		p := Property{}
		if err = p.Read(reader); err != nil {
			return fmt.Errorf("static prop: %w", err)
		}

		e.StaticProperties = append(e.StaticProperties, p)
	}

	for i := 0; i < int(e.Header.DynamicPropertyCount); i++ {
		p := Property{}
		if err = p.Read(reader); err != nil {
			return fmt.Errorf("dynamic prop: %w", err)
		}

		e.DynamicProperties = append(e.DynamicProperties, p)
	}

	return nil
}

func (e *Entity) GetStrings() []string {
	res := []string{e.ClassNameString}
	for _, v := range e.StaticProperties {
		res = append(res, v.NameValue)
		res = append(res, v.Value.GetStrings()...)
	}
	for _, v := range e.DynamicProperties {
		res = append(res, v.NameValue)
		res = append(res, v.Value.GetStrings()...)
	}

	return res
}

func (e *Entity) Write(writer io.WriteSeeker) error {
	var err error
	e.Header.DynamicPropertyCount = uint16(len(e.DynamicProperties))
	e.Header.StaticPropertyCount = uint16(len(e.StaticProperties))
	e.Header.HeaderSize = EntityHeaderSize
	e.Header.ClassNameHash = hashing.StrCode64([]byte(e.ClassNameString))
	e.Header.Magic1 = EntityHeaderMagic
	e.Header.Offset = int32(EntityHeaderSize)

	headerOff, err := writer.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}

	if _, err = writer.Seek(int64(EntityHeaderSize), io.SeekCurrent); err != nil {
		return err
	}

	var dataEnd int64
	var staticSize int64
	{
		//slog.Info("static properties", "len", len(e.StaticProperties))
		for _, p := range e.StaticProperties {
			//slog.Info("pp", "v", fmt.Sprintf("%+v", p))
			if err = p.Write(writer); err != nil {
				return fmt.Errorf("static prop: %w", err)
			}
		}
		staticOff, _ := writer.Seek(0, io.SeekCurrent)
		staticSize = staticOff - headerOff
		for _, p := range e.DynamicProperties {
			if err = p.Write(writer); err != nil {
				return fmt.Errorf("dynamic prop: %w", err)
			}
		}

		if dataEnd, err = writer.Seek(0, io.SeekCurrent); err != nil {
			return err
		}
	}

	dataSize := dataEnd - headerOff
	//slog.Info("datasize", "v", dataSize)
	e.Header.StaticDataSize = int32(staticSize)
	e.Header.DataSize = int32(dataSize)
	//slog.Info("entity header", "v", fmt.Sprintf("%+v", e.Header))
	if _, err = writer.Seek(headerOff, io.SeekStart); err != nil {
		return err
	}
	if err = e.Header.Write(writer); err != nil {
		return fmt.Errorf("header: %w", err)
	}
	_, _ = util.AlignWrite(writer, 16)

	//o, _ := writer.Seek(0, io.SeekCurrent)
	//slog.Info("write entity", "header at", o)

	if _, err = writer.Seek(dataEnd, io.SeekStart); err != nil {
		return err
	}

	return nil
}
