package fpk

import (
	"encoding/binary"
	"fmt"
	"io"
)

type StringHeader struct {
	Offset uint32
	Skip1  uint32
	Length uint32
	Skip2  uint32
}

type String struct {
	Header StringHeader
	Data   string
}

func (sh *StringHeader) Read(reader io.Reader) error {
	return binary.Read(reader, binary.LittleEndian, sh)
}

func (sh *StringHeader) Write(writer io.WriteSeeker) error {
	//o, _ := writer.Seek(0, io.SeekCurrent)
	//slog.Info("stringHeader", "off", o, "len", sh.Length)
	return binary.Write(writer, binary.LittleEndian, sh)
}

func (s *String) Read(reader io.ReadSeeker) error {
	var err error
	if err = s.Header.Read(reader); err != nil {
		return fmt.Errorf("string header: %w", err)
	}

	curPos, err := reader.Seek(0, io.SeekCurrent)
	if err != nil {
		return fmt.Errorf("cannot get current pos: %w", err)
	}

	if _, err = reader.Seek(int64(s.Header.Offset), io.SeekStart); err != nil {
		return fmt.Errorf("seek: %w", err)
	}

	d := make([]byte, s.Header.Length)

	if _, err = reader.Read(d); err != nil {
		return fmt.Errorf("read fpkstring: %w", err)
	}

	_, _ = reader.Seek(curPos, io.SeekStart)

	s.Data = string(d)

	return nil
}

func (s *String) WriteData(writer io.WriteSeeker) error {
	pos, err := writer.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}

	s.Header.Offset = uint32(pos)
	s.Header.Length = uint32(len(s.Data))
	//slog.Info("write string data", "len", s.Header.Length, "offset", pos, "data", s.Data)

	data := append([]byte(s.Data), 0x0)

	if _, err = writer.Write(data); err != nil {
		return err
	}

	return nil
}
