package fox

import (
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type EntityHandle struct {
	Value uint64 `xml:",chardata"`
}

func (eh *EntityHandle) Read(reader io.Reader) error {
	return binary.Read(reader, binary.LittleEndian, &eh.Value)
}

func (eh *EntityHandle) Write(writer io.Writer) error {
	return binary.Write(writer, binary.LittleEndian, eh.Value)
}

func (eh *EntityHandle) String() []string {
	return nil
}

func (eh *EntityHandle) HashString() string {
	return fmt.Sprintf("0x%X", eh.Value)
}

func (eh *EntityHandle) Resolve(m map[uint64]string) {
	return
}

type ehXml struct {
	Value string `xml:",chardata"`
}

func (eh *EntityHandle) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	ss := ehXml{
		Value: fmt.Sprintf("0x%X", eh.Value),
	}
	return e.EncodeElement(ss, start)
}

func (eh *EntityHandle) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var err error
	ss := ehXml{}
	if err = d.DecodeElement(&ss, &start); err != nil {
		return err
	}

	if eh.Value, err = strconv.ParseUint(strings.TrimPrefix(ss.Value, "0x"), 16, 64); err != nil {
		return err
	}

	return nil
}
