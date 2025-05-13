package fox

import (
	"encoding/binary"
	"io"
)

type WideVector3 struct {
	X float32 `xml:"x,attr"`
	Y float32 `xml:"y,attr"`
	Z float32 `xml:"z,attr"`
	A uint16  `xml:"a,attr"`
	B uint16  `xml:"b,attr"`
}

func (i *WideVector3) Read(reader io.Reader) error {
	if err := binary.Read(reader, binary.LittleEndian, i); err != nil {
		return err
	}
	return nil
}

func (i *WideVector3) Write(writer io.Writer) error {
	return binary.Write(writer, binary.LittleEndian, i)
}

func (i *WideVector3) String() []string {
	return nil
}

func (i *WideVector3) HashString() string {
	return ""
}

func (i *WideVector3) Resolve(m map[uint64]string) {
	return
}
