package fox

import (
	"encoding/binary"
	"fmt"
	"io"
)

type UInt8 struct {
	Value uint8 `xml:",chardata"`
}

func (i *UInt8) Read(reader io.Reader) error {
	if err := binary.Read(reader, binary.LittleEndian, &i.Value); err != nil {
		return err
	}
	return nil
}

func (i *UInt8) Write(writer io.Writer) error {
	return binary.Write(writer, binary.LittleEndian, i.Value)
}

func (i *UInt8) String() []string {
	return nil
}

func (i *UInt8) HashString() string {
	return fmt.Sprintf("%d", i.Value)
}

func (i *UInt8) Resolve(m map[uint64]string) {
	return
}
