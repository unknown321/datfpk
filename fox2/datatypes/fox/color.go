package fox

import (
	"encoding/binary"
	"io"
)

type Color struct {
	R float32 `xml:"r,attr"`
	G float32 `xml:"g,attr"`
	B float32 `xml:"b,attr"`
	A float32 `xml:"a,attr"`
}

func (i *Color) Read(reader io.Reader) error {
	if err := binary.Read(reader, binary.LittleEndian, i); err != nil {
		return err
	}
	return nil
}

func (i *Color) Write(writer io.Writer) error {
	return binary.Write(writer, binary.LittleEndian, i)
}

func (i *Color) String() []string {
	return nil
}

func (i *Color) HashString() string {
	return ""
}

func (i *Color) Resolve(m map[uint64]string) {
	return
}
