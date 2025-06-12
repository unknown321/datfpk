package fpk

import (
	"encoding/json"
	"fmt"
	"github.com/unknown321/datfpk/util"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

var MagicFpk = [10]byte{0x66, 0x6f, 0x78, 0x66, 0x70, 0x6b, 0x00, 0x77, 0x69, 0x6e}  // "foxfpk\000win"
var MagicFpkd = [10]byte{0x66, 0x6f, 0x78, 0x66, 0x70, 0x6b, 0x64, 0x77, 0x69, 0x6e} // "foxfpkdwin"
const FpkID = "fpk"
const FpkdID = "fpkd"

type Fpk struct {
	Header     Header
	Entries    []Entry
	References []Reference
	FilePath   string `json:"-"`

	handle io.ReadSeeker
}

type fjs struct {
	Type       string      `json:"type"`
	Entries    []Entry     `json:"entries,omitempty"`
	References []Reference `json:"references,omitempty"`
}

func (f *Fpk) MarshalJSON() ([]byte, error) {
	t := ""
	switch f.Header.Magic {
	case MagicFpk:
		t = FpkID
	case MagicFpkd:
		t = FpkdID
	default:
		return nil, fmt.Errorf("wrong magic % x (%s)", f.Header.Magic, f.Header.Magic)
	}

	//if f.Header.Magic == MagicFpk {
	//	t = FpkID
	//} else {
	//	if f.Header.Magic
	//}
	j := fjs{
		Type:       t,
		Entries:    f.Entries,
		References: f.References,
	}

	return json.Marshal(j)
}

func (f *Fpk) UnmarshalJSON(bytes []byte) error {
	j := &fjs{}
	if err := json.Unmarshal(bytes, j); err != nil {
		return err
	}

	switch j.Type {
	case FpkID:
		f.Header.Magic = MagicFpk
	case FpkdID:
		f.Header.Magic = MagicFpkd
	default:
		return fmt.Errorf("wrong type %s", j.Type)
	}

	f.Entries = j.Entries
	f.References = j.References
	f.Header.MagicNumber2 = 2
	f.Header.EntryCount = uint32(len(j.Entries))
	f.Header.RefCount = uint32(len(j.References))

	return nil
}

func (f *Fpk) ReadFrom(path string, printLog bool) error {
	var file *os.File
	var err error
	f.FilePath = path
	if file, err = os.Open(path); err != nil {
		return fmt.Errorf("cannot open: %w", err)
	}
	//defer file.Close()

	err = f.Read(file, printLog)
	return nil
}

func (f *Fpk) Read(reader io.ReadSeeker, printLog bool) error {
	f.handle = reader

	var err error
	if err = f.Header.Read(reader); err != nil {
		return fmt.Errorf("fpk header: %w", err)
	}

	//o, _ := reader.Seek(0, io.SeekCurrent)
	//slog.Info("entries", "off", o)

	for i := 0; i < int(f.Header.EntryCount); i++ {
		e := Entry{}
		if err = e.Read(reader); err != nil {
			return fmt.Errorf("entry %d read: %w", i, err)
		}
		f.Entries = append(f.Entries, e)
		if printLog {
			slog.Info("entry", "filePath", e.FilePath.Data)
		}
	}

	//o, _ = reader.Seek(0, io.SeekCurrent)
	//slog.Info("ref", "off", o)

	for i := 0; i < int(f.Header.RefCount); i++ {
		r := Reference{}
		if err = r.FilePath.Read(reader); err != nil {
			return fmt.Errorf("reference %d read: %w", i, err)
		}
		f.References = append(f.References, r)
		if printLog {
			slog.Info("reference", "filePath", r.FilePath.Data, "dataOffset", r.FilePath.Header.Offset)
		}
	}

	return nil
}

func (f *Fpk) Write(file io.ReadWriteSeeker, baseDir string, printLog bool) error {
	var err error

	headerSkip := int64(HeaderSize)
	entrySkip := int64(len(f.Entries) * EntrySize)
	refSkip := int64(len(f.References) * ReferenceSize)
	refDataPos := headerSkip + entrySkip + refSkip
	//slog.Info("skip", "header", headerSkip, "entry", entrySkip, "ref", refSkip)

	if _, err = file.Seek(refDataPos, io.SeekStart); err != nil {
		return fmt.Errorf("ref skip: %w", err)
	}

	//o, _ := file.Seek(0, io.SeekCurrent)
	//slog.Info("header string data", "off", o)

	for i := range f.Entries {
		if err = f.Entries[i].FilePath.WriteData(file); err != nil {
			return fmt.Errorf("entry filePath: %w", err)
		}
	}

	//o, _ = file.Seek(0, io.SeekCurrent)
	//slog.Info("reference data", "off", o)

	for i := range f.References {
		if err = f.References[i].FilePath.WriteData(file); err != nil {
			return fmt.Errorf("ref filePath: %w", err)
		}
	}

	if _, err = util.AlignWrite(file, 16); err != nil {
		return err
	}

	//o, _ = file.Seek(0, io.SeekCurrent)
	//slog.Info("write entry data", "offset", o)

	for i := range f.Entries {
		if len(f.Entries[i].Data) == 0 {
			p := filepath.Join(baseDir, f.Entries[i].FilePath.Data)
			if f.Entries[i].Data, err = os.ReadFile(p); err != nil {
				return fmt.Errorf("read entry data: %w", err)
			}
		}

		if err = f.Entries[i].WriteData(file, f.Entries[i].FilePath.Data); err != nil {
			return fmt.Errorf("write entry, path %s, error %w", f.Entries[i].FilePath.Data, err)
		}

		if _, err = util.AlignWrite(file, 16); err != nil {
			return err
		}

		if printLog {
			slog.Info("entry", "path", f.Entries[i].FilePath.Data)
		}
	}

	fSize, _ := file.Seek(0, io.SeekCurrent)
	f.Header.FileSize = uint32(fSize)
	f.Header.RefCount = uint32(len(f.References))
	f.Header.EntryCount = uint32(len(f.Entries))

	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return err
	}

	if err = f.Header.Write(file); err != nil {
		return fmt.Errorf("header write: %w", err)
	}

	//o, _ = file.Seek(0, io.SeekCurrent)
	//slog.Info("entries", "off", o)

	for i := range f.Entries {
		if err = f.Entries[i].WriteHeader(file); err != nil {
			return fmt.Errorf("entry header write: %w", err)
		}
	}

	//o, _ = file.Seek(0, io.SeekCurrent)
	//slog.Info("references", "off", o)

	for _, r := range f.References {
		if err = r.FilePath.Header.Write(file); err != nil {
			return fmt.Errorf("ref header: %w", err)
		}
	}

	return nil
}

func (f *Fpk) SaveDefinition(writer io.Writer) error {
	o, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return fmt.Errorf("save definition: %w", err)
	}

	if _, err = writer.Write(o); err != nil {
		return fmt.Errorf("save definition: %w", err)
	}
	return nil
}

func (f *Fpk) Extract(path string, outDir string) error {
	datDirName := outDir
	workdir := ""
	if outDir == "" {
		workdir = filepath.Dir(f.FilePath)
		ext := filepath.Ext(f.FilePath)
		datDirName = strings.TrimSuffix(filepath.Base(f.FilePath), ext) + strings.ReplaceAll(ext, ".", "_")
	}
	outP := filepath.Dir(path)
	outDir = filepath.Join(workdir, datDirName, outP)
	if err := os.MkdirAll(outDir, os.ModePerm); err != nil {
		return fmt.Errorf("outdir %s: %w", outDir, err)
	}
	outfilePath := filepath.Join(outDir, filepath.Base(path))
	outFile, err := os.OpenFile(outfilePath, os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("extract open output file: %w", err)
	}
	//defer outFile.Close()

	if err = f.ExtractTo(path, outFile); err != nil {
		return fmt.Errorf("extract: %w", err)
	}

	return nil
}
func (f *Fpk) Close() {
	if f.handle != nil {
		f.handle.(*os.File).Close()
	}
}

func (f *Fpk) ExtractTo(path string, outFile io.WriteSeeker) error {
	var e *Entry = nil
	for _, v := range f.Entries {
		if v.FilePath.Data == path {
			e = &v
			break
		}
	}

	if e == nil {
		return fmt.Errorf("entry not found, path %s", path)
	}

	if err := e.ReadData(f.handle); err != nil {
		return fmt.Errorf("fpk entry read data: %w", err)
	}

	if _, err := outFile.Write(e.Data); err != nil {
		return fmt.Errorf("fpk entry extract data: %w", err)
	}

	return nil
}
