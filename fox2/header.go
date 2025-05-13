package fox2

import (
	"encoding/binary"
	"io"
)

var Magic1 uint32 = 0x786f62f2
var Magic2 uint32 = 0x35

type Header struct {
	Magic1            uint32   `xml:"-"`
	Magic2            uint32   `xml:"-"`
	EntityCount       uint32   `xml:"-"`
	StringTableOffset uint32   `xml:"-"`
	DataOffset        uint32   `xml:"-"`
	_                 [12]byte `xml:"-"`
}

const FoxHeaderSize = 32

func (h *Header) Read(reader io.Reader) error {
	return binary.Read(reader, binary.LittleEndian, h)
}

func (h *Header) Write(writer io.Writer) error {
	h.Magic1 = Magic1
	h.Magic2 = Magic2
	return binary.Write(writer, binary.LittleEndian, h)
}
