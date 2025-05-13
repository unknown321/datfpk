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

type FilePtr struct {
	Hash  uint64
	Value string
}

func (f *FilePtr) Read(reader io.Reader) error {
	if err := binary.Read(reader, binary.LittleEndian, &f.Hash); err != nil {
		return err
	}
	return nil
}

func (f *FilePtr) Write(writer io.Writer) error {
	f.Hash = hashing.StrCode64([]byte(f.Value))
	return binary.Write(writer, binary.LittleEndian, f.Hash)
}

func (f *FilePtr) String() []string {
	return []string{f.Value}
}

func (f *FilePtr) HashString() string {
	return fmt.Sprintf("0x%X", f.Hash)
}

func (f *FilePtr) Resolve(m map[uint64]string) {
	var ok bool
	if f.Value, ok = m[f.Hash]; !ok {
		f.Value = ""
	}
}

type fptr struct {
	Hash  string `xml:"hash,attr,omitempty"`
	Value string `xml:",chardata"`
}

func (f *FilePtr) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	ff := fptr{
		Hash:  fmt.Sprintf("0x%X", f.Hash),
		Value: f.Value,
	}

	if ff.Value != "" {
		ff.Hash = ""
	}

	return e.EncodeElement(ff, start)
}

func (f *FilePtr) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var err error
	pp := &fptr{}
	if err = d.DecodeElement(pp, &start); err != nil {
		return err
	}

	if pp.Hash != "" {
		if f.Hash, err = strconv.ParseUint(strings.TrimPrefix(pp.Value, "0x"), 16, 64); err != nil {
			return err
		}
	}

	f.Value = pp.Value
	return nil
}
