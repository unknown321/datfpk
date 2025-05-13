package fox

import (
	"encoding/binary"
	"io"
)

type Vector3 struct {
	X float32 `xml:"x,attr"`
	Y float32 `xml:"y,attr"`
	Z float32 `xml:"z,attr"`
	W float32 `xml:"w,attr"`
}

func (i *Vector3) Read(reader io.Reader) error {
	if err := binary.Read(reader, binary.LittleEndian, i); err != nil {
		return err
	}
	return nil
}

func (i *Vector3) Write(writer io.Writer) error {
	return binary.Write(writer, binary.LittleEndian, i)
}

func (i *Vector3) String() []string {
	return nil
}

func (i *Vector3) HashString() string {
	return ""
}

func (i *Vector3) Resolve(m map[uint64]string) {
	return
}
