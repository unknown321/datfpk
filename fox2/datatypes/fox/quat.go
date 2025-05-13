package fox

import (
	"encoding/binary"
	"io"
)

type Quat struct {
	X float32 `xml:"x,attr"`
	Y float32 `xml:"y,attr"`
	Z float32 `xml:"z,attr"`
	W float32 `xml:"w,attr"`
}

func (q *Quat) Read(reader io.Reader) error {
	if err := binary.Read(reader, binary.LittleEndian, q); err != nil {
		return err
	}
	return nil
}

func (q *Quat) Write(writer io.Writer) error {
	return binary.Write(writer, binary.LittleEndian, q)
}

func (q *Quat) String() []string {
	return nil
}

func (q *Quat) HashString() string {
	return ""
}

func (q *Quat) Resolve(m map[uint64]string) {
	return
}
