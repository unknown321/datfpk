package fox2

import (
	"datfpk/fox2/containers"
	"datfpk/fox2/datatypes/fox"
	"datfpk/hashing"
	"datfpk/util"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"strconv"
)

type PropertyHeader struct {
	NameHash      uint64
	DataType      fox.FDataType
	ContainerType containers.FoxContainerType
	ValueCount    int16
	Offset        int16
	Size          uint16
	Unknown2      int32
	Unknown3      int32
	Unknown4      int32
	Unknown5      int32
}

var PropertyHeaderSize int16 = 32

func (ph *PropertyHeader) Read(reader io.Reader) error {
	return binary.Read(reader, binary.LittleEndian, ph)
}

func (ph *PropertyHeader) Write(writer io.Writer) error {
	return binary.Write(writer, binary.LittleEndian, ph)
}

type Property struct {
	Header PropertyHeader
	Value  IFoxContainer

	NameValue string
}

func (p *Property) Read(reader io.ReadSeeker) error {
	var err error
	//o, _ := reader.Seek(0, io.SeekCurrent)
	//slog.Info("prop header", "read", o)

	if err = p.Header.Read(reader); err != nil {
		return fmt.Errorf("property header: %w", err)
	}
	//slog.Info("property", "nameHash", fmt.Sprintf("%x", p.Header.NameHash), "datatype", p.Header.DataType,
	//	"containerType", p.Header.ContainerType, "value count", p.Header.ValueCount, "offset", p.Header.Offset,
	//	"size", p.Header.Size)

	//o, _ = reader.Seek(0, io.SeekCurrent)
	//slog.Info("prop container", "read", o)

	if err = ReadContainer(&p.Value, reader, p.Header.DataType, p.Header.ContainerType, p.Header.ValueCount); err != nil {
		//if err = p.Value.Read(reader, p.Header.DataType, p.Header.ContainerType, p.Header.ValueCount); err != nil {
		return fmt.Errorf("container: %w", err)
	}

	if p.Value == nil {
		panic("property value is nil")
	}

	if _, err = util.AlignRead(reader, 16); err != nil {
		return err
	}

	return nil
}

func (p *Property) Write(writer io.WriteSeeker) error {
	var err error
	var headerOffset int64
	//var pos int64
	if headerOffset, err = writer.Seek(0, io.SeekCurrent); err != nil {
		return err
	}
	if _, err = writer.Seek(int64(PropertyHeaderSize), io.SeekCurrent); err != nil {
		return err
	}
	//slog.Info("write prop value", "pos", pos)
	if err = p.Value.Write(writer); err != nil {
		return fmt.Errorf("container: %w", err)
	}
	_, _ = util.AlignWrite(writer, 16)

	dataEnd, err := writer.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}

	if _, err = writer.Seek(headerOffset, io.SeekStart); err != nil {
		return err
	}
	p.Header.NameHash = hashing.StrCode64([]byte(p.NameValue))
	p.Header.Offset = PropertyHeaderSize
	p.Header.Size = uint16(dataEnd - headerOffset)
	//slog.Info("property header size", "s", p.Header.Size)
	if err = p.Header.Write(writer); err != nil {
		return fmt.Errorf("header: %w", err)
	}

	_, _ = writer.Seek(dataEnd, io.SeekStart)

	return nil
}

type pXml struct {
	XMLName   xml.Name      `xml:"property"`
	Name      string        `xml:"name,attr"`
	Type      string        `xml:"type,attr"`
	Container string        `xml:"container,attr"`
	ArraySize int16         `xml:"arraySize,attr,omitempty"`
	Value     IFoxContainer `xml:"containerEntry"`
	Unknown2  int32         `xml:"unknown2,attr,omitempty"`
	Unknown3  int32         `xml:"unknown3,attr,omitempty"`
	Unknown4  int32         `xml:"unknown4,attr,omitempty"`
	Unknown5  int32         `xml:"unknown5,attr,omitempty"`
}

func (p *Property) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	px := pXml{
		Name:      p.NameValue,
		Type:      fox.DataTypeToString(p.Header.DataType),
		Container: p.Header.ContainerType.String(),
		ArraySize: p.Header.ValueCount,
		Value:     p.Value,
		Unknown2:  p.Header.Unknown2,
		Unknown3:  p.Header.Unknown3,
		Unknown4:  p.Header.Unknown4,
		Unknown5:  p.Header.Unknown5,
	}

	if px.Value == nil {
		return fmt.Errorf("value is nil")
	}

	return e.Encode(px)
}

func (p *Property) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var err error

	for _, a := range start.Attr {
		//fmt.Printf("%s = %s\n", a.Name.Local, a.Value)
		switch a.Name.Local {
		case "name":
			p.NameValue = a.Value
		case "type":
			if p.Header.DataType, err = fox.DataTypeFromString(a.Value); err != nil {
				return err
			}
		case "container":
			if p.Header.ContainerType, err = containers.ContainerTypeFromString(a.Value); err != nil {
				return err
			}
		case "arraySize":
			val, err := strconv.ParseInt(a.Value, 10, 16)
			if err != nil {
				return err
			}
			p.Header.ValueCount = int16(val)
		case "unknown2":
			val, err := strconv.ParseInt(a.Value, 10, 32)
			if err != nil {
				return err
			}
			p.Header.Unknown2 = int32(val)
		case "unknown3":
			val, err := strconv.ParseInt(a.Value, 10, 32)
			if err != nil {
				return err
			}
			p.Header.Unknown3 = int32(val)
		case "unknown4":
			val, err := strconv.ParseInt(a.Value, 10, 32)
			if err != nil {
				return err
			}
			p.Header.Unknown4 = int32(val)
		case "unknown5":
			val, err := strconv.ParseInt(a.Value, 10, 32)
			if err != nil {
				return err
			}
			p.Header.Unknown5 = int32(val)
		default:
			slog.Warn("ignored property attribute %s, value", a.Name.Local, a.Value)
		}

	}

	var cont IFoxContainer
	if cont, err = CreateTypedContainer(p.Header.DataType, p.Header.ContainerType, int(p.Header.ValueCount)); err != nil {
		return err
	}

	p.Value = cont

	for {
		t, err := d.Token()
		if err != nil {
			panic(err)
		}

		switch tt := t.(type) {
		case xml.StartElement:
			//fmt.Printf("<%s> %+v\n", tt.Name.Local, tt.Attr)
			if err = p.Value.DecodeNext(d, &tt); err != nil {
				return err
			}
		//case xml.CharData:
		//	fmt.Println("chardata")
		//	fmt.Printf("->%s\n", tt)
		case xml.EndElement:
			//fmt.Println("end")
			//fmt.Printf("</%s>\n", tt.Name.Local)
			if tt == start.End() {
				return nil
			}
		}
	}

	return nil
}
