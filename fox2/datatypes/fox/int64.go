package fox

import (
	"encoding/binary"
	"fmt"
	"io"
)

type Int64 struct {
	Value int64 `xml:",chardata"`
}

func (i *Int64) Read(reader io.Reader) error {
	if err := binary.Read(reader, binary.LittleEndian, &i.Value); err != nil {
		return err
	}
	return nil
}

func (i *Int64) Write(writer io.Writer) error {
	return binary.Write(writer, binary.LittleEndian, i.Value)
}

func (i *Int64) String() []string {
	return nil
}

func (i *Int64) HashString() string {
	return fmt.Sprintf("%d", i.Value)
}

func (i *Int64) Resolve(m map[uint64]string) {
	return
}
