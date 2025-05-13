package fox

import (
	"encoding/binary"
	"fmt"
	"io"
)

type UInt16 struct {
	Value uint16 `xml:",chardata"`
}

func (i *UInt16) Read(reader io.Reader) error {
	if err := binary.Read(reader, binary.LittleEndian, &i.Value); err != nil {
		return err
	}
	return nil
}

func (i *UInt16) Write(writer io.Writer) error {
	return binary.Write(writer, binary.LittleEndian, i.Value)
}

func (i *UInt16) String() []string {
	return nil
}

func (i *UInt16) HashString() string {
	return fmt.Sprintf("%d", i.Value)
}

func (i *UInt16) Resolve(m map[uint64]string) {
	return
}
