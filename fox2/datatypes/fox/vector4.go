package fox

import (
	"encoding/binary"
	"io"
)

type Vector4 struct {
	X float32 `xml:"x,attr"`
	Y float32 `xml:"y,attr"`
	Z float32 `xml:"z,attr"`
	W float32 `xml:"w,attr"`
}

func (i *Vector4) Read(reader io.Reader) error {
	if err := binary.Read(reader, binary.LittleEndian, i); err != nil {
		return err
	}
	return nil
}

func (i *Vector4) Write(writer io.Writer) error {
	return binary.Write(writer, binary.LittleEndian, i)
}

func (i *Vector4) String() []string {
	return nil
}

func (i *Vector4) HashString() string {
	return ""
}

func (i *Vector4) Resolve(m map[uint64]string) {
	return
}
