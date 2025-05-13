package fox

import (
	"encoding/binary"
	"fmt"
	"io"
)

type Int32 struct {
	Value int32 `xml:",chardata"`
}

func (i *Int32) Read(reader io.Reader) error {
	if err := binary.Read(reader, binary.LittleEndian, &i.Value); err != nil {
		return err
	}
	return nil
}

func (i *Int32) Write(writer io.Writer) error {
	return binary.Write(writer, binary.LittleEndian, i.Value)
}

func (i *Int32) String() []string {
	return nil
}

func (i *Int32) HashString() string {
	return fmt.Sprintf("%d", i.Value)
}

func (i *Int32) Resolve(m map[uint64]string) {
	return
}
