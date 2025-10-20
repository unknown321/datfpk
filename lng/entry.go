package lng

import (
	"encoding/binary"
	"fmt"
	"io"
)

type Entry struct {
	LangId string
	Offset int64
	Color  int16
	Value  string
	Key    uint32
}

func (e *Entry) Read(seeker io.ReadSeeker, order binary.ByteOrder) error {
	var err error

	b := make([]byte, 1)

	var value []byte

	if e.Offset, err = seeker.Seek(0, io.SeekCurrent); err != nil {
		return fmt.Errorf("reading offset: %w", err)
	}

	// TODO is color always le?
	if err = binary.Read(seeker, binary.LittleEndian, &e.Color); err != nil {
		return fmt.Errorf("color: %w", err)
	}

	for {
		n, err := seeker.Read(b)
		if err != nil {
			return fmt.Errorf("value: %w", err)
		}

		if n == 0 {
			return fmt.Errorf("read: too few bytes")
		}

		if b[0] == 0 {
			break
		}

		value = append(value, b[0])
	}

	e.Value = string(value)

	return nil
}

func (e *Entry) Write(w io.Writer, order binary.ByteOrder) error {
	var err error
	if err = binary.Write(w, binary.LittleEndian, e.Color); err != nil {
		return fmt.Errorf("color: %w", err)
	}

	if _, err = w.Write([]byte(e.Value)); err != nil {
		return fmt.Errorf("value: %w", err)
	}

	if _, err = w.Write([]byte{0}); err != nil {
		return fmt.Errorf("nullbyte: %w", err)
	}

	return nil
}
