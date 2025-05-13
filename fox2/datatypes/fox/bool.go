package fox

import (
	"encoding/binary"
	"io"
)

type Bool struct {
	Value bool `xml:",chardata"`
}

func (b *Bool) Read(reader io.Reader) error {
	if err := binary.Read(reader, binary.LittleEndian, &b.Value); err != nil {
		return err
	}
	return nil
}

func (b *Bool) Write(writer io.Writer) error {
	return binary.Write(writer, binary.LittleEndian, b.Value)
}

func (b *Bool) Resolve(m map[uint64]string) {
	return
}

func (b *Bool) String() []string {
	return nil
}
