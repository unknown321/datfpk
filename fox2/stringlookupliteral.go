package fox2

import (
	"datfpk/hashing"
	"encoding/binary"
	"io"
)

type StringLookupLiteral struct {
	Hash      uint64
	Length    int32
	Literal   string
	Encrypted []byte
}

func (s *StringLookupLiteral) Read(reader io.ReadSeeker) bool {
	var err error
	if err = binary.Read(reader, binary.LittleEndian, &s.Hash); err != nil {
		return false
	}

	if s.Hash == 0 {
		return false
	}

	if err = binary.Read(reader, binary.LittleEndian, &s.Length); err != nil {
		return false
	}

	l := make([]byte, s.Length)
	if _, err = reader.Read(l); err != nil {
		return false
	}

	s.Literal = string(l)

	return true
}

func (s *StringLookupLiteral) Write(writer io.Writer) error {
	var err error

	if s.Hash == 0 {
		s.Hash = hashing.StrCode64([]byte(s.Literal))
	}

	s.Length = int32(len(s.Literal))

	if err = binary.Write(writer, binary.LittleEndian, s.Hash); err != nil {
		return err
	}
	if err = binary.Write(writer, binary.LittleEndian, s.Length); err != nil {
		return err
	}
	if _, err = writer.Write([]byte(s.Literal)); err != nil {
		return err
	}

	return nil
}
