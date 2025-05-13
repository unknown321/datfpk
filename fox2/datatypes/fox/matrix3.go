package fox

import (
	"encoding/binary"
	"io"
)

type Matrix3 struct {
	R11 float32 `xml:"r11,attr"`
	R12 float32 `xml:"r12,attr"`
	R13 float32 `xml:"r13,attr"`
	R21 float32 `xml:"r21,attr"`
	R22 float32 `xml:"r22,attr"`
	R23 float32 `xml:"r23,attr"`
	R31 float32 `xml:"r31,attr"`
	R32 float32 `xml:"r32,attr"`
	R33 float32 `xml:"r33,attr"`
}

func (i *Matrix3) Read(reader io.Reader) error {
	if err := binary.Read(reader, binary.LittleEndian, i); err != nil {
		return err
	}
	return nil
}

func (i *Matrix3) Write(writer io.Writer) error {
	return binary.Write(writer, binary.LittleEndian, i)
}

func (i *Matrix3) String() []string {
	return nil
}

func (i *Matrix3) HashString() string {
	return ""
}

func (i *Matrix3) Resolve(m map[uint64]string) {
	return
}
