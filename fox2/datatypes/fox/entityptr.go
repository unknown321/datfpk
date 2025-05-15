package fox

import (
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type EntityPtr struct {
	Value uint64
}

func (ep *EntityPtr) Read(reader io.Reader) error {
	return binary.Read(reader, binary.LittleEndian, &ep.Value)
}

func (ep *EntityPtr) Write(writer io.Writer) error {
	return binary.Write(writer, binary.LittleEndian, ep.Value)
}

type epXml struct {
	Value string `xml:",chardata"`
}

func (ep *EntityPtr) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	ss := epXml{
		Value: fmt.Sprintf("0x%X", ep.Value),
	}
	return e.EncodeElement(ss, start)
}

func (ep *EntityPtr) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var err error
	ss := epXml{}
	if err = d.DecodeElement(&ss, &start); err != nil {
		return err
	}

	if ep.Value, err = strconv.ParseUint(strings.TrimPrefix(ss.Value, "0x"), 16, 64); err != nil {
		return fmt.Errorf("entityPtr: %w", err)
	}

	return nil
}

func (ep *EntityPtr) Resolve(m map[uint64]string) {
	return
}

func (ep *EntityPtr) String() []string {
	return nil
}
