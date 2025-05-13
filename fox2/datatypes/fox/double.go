package fox

import (
	"encoding/binary"
	"fmt"
	"io"
)

type Double struct {
	Value float64 `xml:",chardata"`
}

func (i *Double) Read(reader io.Reader) error {
	if err := binary.Read(reader, binary.LittleEndian, &i.Value); err != nil {
		return err
	}

	return nil
}

func (i *Double) Write(writer io.Writer) error {
	return binary.Write(writer, binary.LittleEndian, i.Value)
}

func (i *Double) String() []string {
	return nil
}

func (i *Double) HashString() string {
	return fmt.Sprintf("%f", i.Value)
}

func (i *Double) Resolve(m map[uint64]string) {
	return
}
