package fox

import (
	"encoding/binary"
	"io"
)

type Matrix4 struct {
	R11 float32 `xml:"r11,attr"`
	R12 float32 `xml:"r12,attr"`
	R13 float32 `xml:"r13,attr"`
	R14 float32 `xml:"r14,attr"`
	R21 float32 `xml:"r21,attr"`
	R22 float32 `xml:"r22,attr"`
	R23 float32 `xml:"r23,attr"`
	R24 float32 `xml:"r24,attr"`
	R31 float32 `xml:"r31,attr"`
	R32 float32 `xml:"r32,attr"`
	R33 float32 `xml:"r33,attr"`
	R34 float32 `xml:"r34,attr"`
	R41 float32 `xml:"r41,attr"`
	R42 float32 `xml:"r42,attr"`
	R43 float32 `xml:"r43,attr"`
	R44 float32 `xml:"r44,attr"`
}

func (i *Matrix4) Read(reader io.Reader) error {
	if err := binary.Read(reader, binary.LittleEndian, i); err != nil {
		return err
	}
	return nil
}

func (i *Matrix4) Write(writer io.Writer) error {
	return binary.Write(writer, binary.LittleEndian, i)
}

func (i *Matrix4) String() []string {
	return nil
}

func (i *Matrix4) HashString() string {
	return ""
}

func (i *Matrix4) Resolve(m map[uint64]string) {
	return
}
