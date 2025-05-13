package fox

import (
	"encoding/binary"
	"fmt"
	"io"
)

type Float struct {
	Value float32 `xml:",chardata"`
}

func (i *Float) Read(reader io.Reader) error {
	if err := binary.Read(reader, binary.LittleEndian, &i.Value); err != nil {
		return err
	}

	return nil
}

func (i *Float) Write(writer io.Writer) error {
	return binary.Write(writer, binary.LittleEndian, i.Value)
}

func (i *Float) String() []string {
	return nil
}

func (i *Float) HashString() string {
	return fmt.Sprintf("%f", i.Value)
}

func (i *Float) Resolve(m map[uint64]string) {
	return
}
