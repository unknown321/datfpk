package lng

import (
	"encoding/binary"
	"io"
)

type Key struct {
	Key    uint32
	Offset uint32
}

func (k *Key) Read(seeker io.ReadSeeker, order binary.ByteOrder) error {
	return binary.Read(seeker, order, k)
}

func (k *Key) Write(writer io.Writer, order binary.ByteOrder) error {
	return binary.Write(writer, order, k)
}
