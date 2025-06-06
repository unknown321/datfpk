package fox

import (
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/unknown321/hashing"
)

type Path struct {
	Hash  uint64
	Value string
}

func (p *Path) Read(reader io.Reader) error {
	if err := binary.Read(reader, binary.LittleEndian, &p.Hash); err != nil {
		return err
	}
	return nil
}

func (p *Path) Write(writer io.Writer) error {
	p.Hash = hashing.StrCode64([]byte(p.Value))
	return binary.Write(writer, binary.LittleEndian, p.Hash)
}

func (p *Path) String() []string {
	return []string{p.Value}
}

func (p *Path) HashString() string {
	return fmt.Sprintf("0x%X", p.Hash)
}

func (p *Path) Resolve(m map[uint64]string) {
	var ok bool
	if p.Value, ok = m[p.Hash]; !ok {
		p.Value = ""
	}
}

type pXml struct {
	Hash  string `xml:"hash,attr,omitempty"`
	Value string `xml:",chardata"`
}

func (p *Path) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	pp := pXml{
		Hash:  fmt.Sprintf("0x%X", p.Hash),
		Value: p.Value,
	}

	if p.Value != "" {
		pp.Hash = ""
	}

	return e.EncodeElement(pp, start)
}

func (p *Path) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var err error
	pp := &pXml{}
	if err = d.DecodeElement(pp, &start); err != nil {
		return err
	}

	if pp.Hash != "" {
		if p.Hash, err = strconv.ParseUint(strings.TrimPrefix(pp.Hash, "0x"), 16, 64); err != nil {
			return fmt.Errorf("Path: %w", err)
		}
	}

	p.Value = pp.Value
	return nil
}
