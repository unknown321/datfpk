package fox2

import (
	"datfpk/hashing"
	"datfpk/util"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"io"
)

type Fox2 struct {
	FormatVersion        int      `xml:"formatVersion,attr"`
	FileVersion          int      `xml:"fileVersion,attr"`
	Header               Header   `xml:"-"`
	Entities             []Entity `xml:"entities>entity"`
	Classes              []Class
	StringLookupLiterals []StringLookupLiteral `xml:"-"`
}

var Trailer = []byte{0x00, 0x00, 0x65, 0x6E, 0x64} // 0 0 end

var fox2dict = make(map[uint64]string)

//func Init(dictPath string) error {
//	if dictPath == "" {
//		dictPath = "fox_dictionary.txt"
//	}
//
//	fox2dict = make(map[uint64]string)
//
//		f, err := os.ReadFile(dictPath)
//		if err != nil {
//			return fmt.Errorf("fox dictionary: %w", err)
//		}
//
//		for _, line := range bytes.Split(f, []byte("\n")) {
//			qq := bytes.TrimSuffix(line, []byte("\r"))
//			v := hashing.StrCode64(qq)
//			fox2dict[v] = string(qq)
//		}
//
//		return nil
//}

type Class struct {
}

func (f *Fox2) Read(reader io.ReadSeeker) error {
	var err error
	if err = f.Header.Read(reader); err != nil {
		return fmt.Errorf("header: %w", err)
	}

	//slog.Info("entity", "count", f.Header.EntityCount, "stringTableOffset", f.Header.StringTableOffset)

	for i := 0; i < int(f.Header.EntityCount); i++ {
		//o, _ := reader.Seek(0, io.SeekCurrent)
		//slog.Info("read entity", "offset", o)
		e := Entity{}
		if err = e.Read(reader); err != nil {
			return fmt.Errorf("read entry: %w", err)
		}

		f.Entities = append(f.Entities, e)
	}

	for {
		s := StringLookupLiteral{}
		if !s.Read(reader) {
			break
		}

		//slog.Info("lookup literal", "name", s.Literal, "hash", fmt.Sprintf("%x", s.Hash))

		f.StringLookupLiterals = append(f.StringLookupLiterals, s)
	}

	resolveMap := make(map[uint64]string)
	for _, ss := range f.StringLookupLiterals {
		resolveMap[ss.Hash] = ss.Literal
	}

	for i := range f.Entities {
		f.Entities[i].Resolve(resolveMap)
	}

	return nil
}

func (f *Fox2) ToXML(writer io.Writer) error {
	var err error
	d, err := xml.MarshalIndent(f, "", "  ")
	if err != nil {
		return fmt.Errorf("fox2 marshal: %w", err)
	}

	if _, err = writer.Write(d); err != nil {
		return fmt.Errorf("fox2 write: %w", err)
	}

	return nil
}

func (f *Fox2) FromXML(reader io.Reader) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	return xml.Unmarshal(data, &f)
}

func (f *Fox2) CollectLiterals() {
	ss := []string{}
	f.StringLookupLiterals = make([]StringLookupLiteral, 0)
	for _, e := range f.Entities {
		ss = append(ss, e.GetStrings()...)
	}

	//sort.SliceStable(ss, func(i, j int) bool {
	//	return ss[i] < ss[j]
	//})

	//ls := util.CompactStringSlice(ss)
	ls := []string{}
Outer:
	for _, s := range ss {
		for _, v := range ls {
			if v == s {
				continue Outer
			}
		}
		ls = append(ls, s)
	}

	for _, s := range ls {
		f.StringLookupLiterals = append(f.StringLookupLiterals, StringLookupLiteral{
			Hash:      hashing.StrCode64([]byte(s)),
			Length:    int32(len(s)),
			Literal:   s,
			Encrypted: nil,
		})
		//slog.Info("new string", "v", s)
	}
}

func (f *Fox2) Write(writer io.WriteSeeker) error {
	var err error

	if _, err = writer.Seek(FoxHeaderSize, io.SeekCurrent); err != nil {
		return err
	}

	for _, e := range f.Entities {
		//o, _ := writer.Seek(0, io.SeekCurrent)
		//slog.Info("write entity", "offset", o)
		if err = e.Write(writer); err != nil {
			return fmt.Errorf("entity: %w", err)
		}
	}

	strOffset, err := writer.Seek(0, io.SeekCurrent)
	if err != nil {
		return fmt.Errorf("strOffset: %w", err)
	}
	f.Header.StringTableOffset = uint32(strOffset)

	f.CollectLiterals()

	for _, s := range f.StringLookupLiterals {
		if err = s.Write(writer); err != nil {
			return fmt.Errorf("stringLiteral: %w", err)
		}
	}

	binary.Write(writer, binary.LittleEndian, int64(0))
	util.AlignWrite(writer, 16)
	writer.Write(Trailer)
	util.AlignWrite(writer, 16)

	f.Header.EntityCount = uint32(len(f.Entities))
	f.Header.DataOffset = FoxHeaderSize
	_, _ = writer.Seek(0, io.SeekStart)
	//slog.Info("string off", "v", f.Header.StringTableOffset)
	if err = f.Header.Write(writer); err != nil {
		return fmt.Errorf("header: %w", err)
	}

	return nil
}

type fox2xml struct {
	XMLName     xml.Name `xml:"fox"`
	Version     int      `xml:"formatVersion,attr"`
	FileVersion int      `xml:"fileVersion,attr"`
	//Classes     []Class  `xml:"classes"`
	Entities []Entity `xml:"entities>entity"`
}

func (f *Fox2) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	ff := fox2xml{
		XMLName: xml.Name{
			Local: "fox",
		},
		Version:     2,
		FileVersion: f.FileVersion,
		//Classes:     f.Classes,
		Entities: f.Entities,
	}

	return e.Encode(ff)
}
