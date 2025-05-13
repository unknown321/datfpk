package fox

import (
	"datfpk/hashing"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type String struct {
	Hash  uint64 `xml:"hash,attr,omitempty"`
	Value string `xml:",chardata"`
}

func (s *String) Read(reader io.Reader) error {
	if err := binary.Read(reader, binary.LittleEndian, &s.Hash); err != nil {
		return err
	}
	return nil
}

func (s *String) Write(writer io.Writer) error {
	s.Hash = hashing.StrCode64([]byte(s.Value))
	return binary.Write(writer, binary.LittleEndian, s.Hash)
}

type sXml struct {
	Value string `xml:",chardata"`
	Hash  string `xml:"hash,attr,omitempty"`
}

func (s *String) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	ss := sXml{
		Value: s.Value,
		Hash:  fmt.Sprintf("0x%X", s.Hash),
	}
	if s.Value != "" {
		ss.Hash = ""
	}
	return e.EncodeElement(ss, start)
}

func (s *String) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	ss := &sXml{}
	if err := d.DecodeElement(ss, &start); err != nil {
		return err
	}

	if ss.Value != "" {
		s.Value = ss.Value
		s.Hash = 0
		return nil
	}

	if ss.Hash != "" {
		v, err := strconv.ParseInt(strings.TrimPrefix(ss.Hash, "0x"), 16, 64)
		if err != nil {
			return err
		}

		s.Hash = uint64(v)
		s.Value = ""
		return nil
	}

	return nil
}

func (s *String) String() []string {
	return []string{s.Value}
}

func (s *String) HashString() string {
	return fmt.Sprintf("0x%X", s.Hash)
}

func (s *String) Resolve(m map[uint64]string) {
	var ok bool
	if s.Value, ok = m[s.Hash]; !ok {
		s.Value = ""
	}
}
