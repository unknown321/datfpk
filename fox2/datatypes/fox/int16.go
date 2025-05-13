package fox

import (
	"encoding/binary"
	"fmt"
	"io"
)

type Int16 struct {
	Value int16 `xml:",chardata"`
}

func (i *Int16) Read(reader io.Reader) error {
	if err := binary.Read(reader, binary.LittleEndian, &i.Value); err != nil {
		return err
	}
	return nil
}

func (i *Int16) Write(writer io.Writer) error {
	return binary.Write(writer, binary.LittleEndian, i.Value)
}

func (i *Int16) String() []string {
	return nil
}

func (i *Int16) HashString() string {
	return fmt.Sprintf("%d", i.Value)
}

func (i *Int16) Resolve(m map[uint64]string) {
	return
}
