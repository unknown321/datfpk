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

type EntityLink struct {
	PackagePathHash   uint64
	PackagePath       string
	ArchivePathHash   uint64
	ArchivePath       string
	NameInArchiveHash uint64
	NameInArchive     string
	EntityHandle      uint64
}

func (el *EntityLink) Read(reader io.Reader) error {
	var err error
	if err = binary.Read(reader, binary.LittleEndian, &el.PackagePathHash); err != nil {
		return err
	}
	if err = binary.Read(reader, binary.LittleEndian, &el.ArchivePathHash); err != nil {
		return err
	}
	if err = binary.Read(reader, binary.LittleEndian, &el.NameInArchiveHash); err != nil {
		return err
	}
	if err = binary.Read(reader, binary.LittleEndian, &el.EntityHandle); err != nil {
		return err
	}

	return nil
}

func (el *EntityLink) Write(writer io.Writer) error {
	var err error

	el.PackagePathHash = hashing.StrCode64([]byte(el.PackagePath))
	el.NameInArchiveHash = hashing.StrCode64([]byte(el.NameInArchive))
	el.ArchivePathHash = hashing.StrCode64([]byte(el.ArchivePath))

	if err = binary.Write(writer, binary.LittleEndian, el.PackagePathHash); err != nil {
		return err
	}
	if err = binary.Write(writer, binary.LittleEndian, el.ArchivePathHash); err != nil {
		return err
	}
	if err = binary.Write(writer, binary.LittleEndian, el.NameInArchiveHash); err != nil {
		return err
	}
	if err = binary.Write(writer, binary.LittleEndian, el.EntityHandle); err != nil {
		return err
	}

	return nil
}

func (el *EntityLink) String() []string {
	res := make([]string, 0)
	if el.NameInArchive != "" {
		res = append(res, el.NameInArchive)
	}
	if el.PackagePath != "" {
		res = append(res, el.PackagePath)
	}
	if el.ArchivePath != "" {
		res = append(res, el.ArchivePath)
	}

	return res
}

func (el *EntityLink) HashString() string {
	return fmt.Sprintf("0x%X", el.EntityHandle)
}

func (el *EntityLink) Resolve(m map[uint64]string) {
	el.PackagePath = m[el.PackagePathHash]
	el.NameInArchive = m[el.NameInArchiveHash]
	el.ArchivePath = m[el.ArchivePathHash]
}

type elXml struct {
	PackagePathHash   string `xml:"packagePathHash,attr,omitempty"`
	PackagePath       string `xml:"packagePath,attr,omitempty"`
	ArchivePathHash   string `xml:"archivePathHash,attr,omitempty"`
	ArchivePath       string `xml:"archivePath,attr,omitempty"`
	NameInArchiveHash string `xml:"nameInArchiveHash,attr,omitempty"`
	NameInArchive     string `xml:"nameInArchive,attr,omitempty"`
	EntityHandle      string `xml:",chardata"`
}

func (el *EntityLink) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	ex := elXml{
		PackagePathHash:   fmt.Sprintf("0x%X", el.PackagePathHash),
		PackagePath:       el.PackagePath,
		ArchivePathHash:   fmt.Sprintf("0x%X", el.ArchivePathHash),
		ArchivePath:       el.ArchivePath,
		NameInArchiveHash: fmt.Sprintf("0x%X", el.NameInArchiveHash),
		NameInArchive:     el.NameInArchive,
		EntityHandle:      fmt.Sprintf("0x%X", el.EntityHandle),
	}

	if ex.PackagePath != "" {
		ex.PackagePathHash = ""
	}

	if ex.ArchivePath != "" {
		ex.ArchivePathHash = ""
	}

	if ex.NameInArchive != "" {
		ex.NameInArchiveHash = ""
	}

	return e.EncodeElement(ex, start)
}

func (el *EntityLink) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var err error
	pp := &elXml{}
	if err = d.DecodeElement(pp, &start); err != nil {
		return err
	}

	if pp.ArchivePathHash != "" {
		if el.ArchivePathHash, err = strconv.ParseUint(strings.TrimPrefix(pp.ArchivePathHash, "0x"), 16, 64); err != nil {
			return fmt.Errorf("entityLink, archivePathHash: %w", err)
		}
	}
	if pp.NameInArchiveHash != "" {
		if el.NameInArchiveHash, err = strconv.ParseUint(strings.TrimPrefix(pp.NameInArchiveHash, "0x"), 16, 64); err != nil {
			return fmt.Errorf("entityLink, nameInArchiveHash: %w", err)
		}
	}
	if pp.PackagePathHash != "" {
		if el.PackagePathHash, err = strconv.ParseUint(strings.TrimPrefix(pp.PackagePathHash, "0x"), 16, 64); err != nil {
			return fmt.Errorf("entityLink, packagePathHash: %w", err)
		}
	}

	el.ArchivePath = pp.ArchivePath
	el.PackagePath = pp.PackagePath
	el.NameInArchive = pp.NameInArchive

	if pp.EntityHandle != "" {
		if el.EntityHandle, err = strconv.ParseUint(strings.TrimPrefix(pp.EntityHandle, "0x"), 16, 64); err != nil {
			return fmt.Errorf("entityHandle: %w", err)
		}
	}

	return nil
}
