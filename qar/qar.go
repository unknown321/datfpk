package qar

import (
	"bytes"
	"datfpk/util"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/unknown321/hashing"
)

type Qar struct {
	Magic [4]byte `json:"-"`
	Flags uint32  `json:"flags"`

	FileCount       uint32 `json:"-"`
	UnknownCount    uint32 `json:"-"`
	BlockFileEnd    uint32 `json:"-"`
	OffsetFirstFile uint32 `json:"-"`
	Version         uint32 `json:"version"`
	Unknown2        uint32 `json:"-"`

	FilePath string  `json:"-"`
	Entries  []Entry `json:"entries"`

	handle io.ReadSeeker
}

var magic = [4]byte{0x53, 0x51, 0x41, 0x52} // SQAR
const QarID = "qar"

type qjs struct {
	Type    string  `json:"type"`
	Flags   uint32  `json:"flags"`
	Version uint32  `json:"version"`
	Entries []Entry `json:"entries"`
}

func (q *Qar) MarshalJSON() ([]byte, error) {
	return json.Marshal(qjs{
		Type:    QarID,
		Flags:   q.Flags,
		Version: q.Version,
		Entries: q.Entries,
	})
}

func (q *Qar) UnmarshalJSON(i []byte) error {
	qq := qjs{}
	if err := json.Unmarshal(i, &qq); err != nil {
		return err
	}

	if qq.Type != QarID {
		return fmt.Errorf("unexpected type %s, want %s", qq.Type, QarID)
	}

	q.Flags = qq.Flags
	q.Entries = qq.Entries
	q.Version = qq.Version

	return nil
}

func (q *Qar) ReadFrom(path string) error {
	var f *os.File
	var err error
	q.FilePath = path
	if f, err = os.Open(path); err != nil {
		return fmt.Errorf("cannot open: %w", err)
	}

	err = q.Read(f)
	return err
}

func (q *Qar) Close() {
	if q.handle != nil {
		q.handle.(*os.File).Close()
	}
}

func (q *Qar) Read(f io.ReadSeeker) error {
	q.handle = f

	var err error
	if err = binary.Read(f, binary.LittleEndian, &q.Magic); err != nil {
		return fmt.Errorf("cannot read magic: %w", err)
	}

	if bytes.Compare(q.Magic[:], magic[:]) != 0 {
		return fmt.Errorf("magic mismatch")
	}

	if err = binary.Read(f, binary.LittleEndian, &q.Flags); err != nil {
		return fmt.Errorf("cannot read flags: %w", err)
	}

	if err = binary.Read(f, binary.LittleEndian, &q.FileCount); err != nil {
		return fmt.Errorf("cannot read file count: %w", err)
	}

	if err = binary.Read(f, binary.LittleEndian, &q.UnknownCount); err != nil {
		return fmt.Errorf("cannot read unknown count: %w", err)
	}

	if err = binary.Read(f, binary.LittleEndian, &q.BlockFileEnd); err != nil {
		return fmt.Errorf("cannot read block file end: %w", err)
	}

	if err = binary.Read(f, binary.LittleEndian, &q.OffsetFirstFile); err != nil {
		return fmt.Errorf("cannot read offset first file: %w", err)
	}

	if err = binary.Read(f, binary.LittleEndian, &q.Version); err != nil {
		return fmt.Errorf("cannot read version: %w", err)
	}

	if err = binary.Read(f, binary.LittleEndian, &q.Unknown2); err != nil {
		return fmt.Errorf("cannot read unknown2: %w", err)
	}

	q.Flags ^= xorMask1
	q.FileCount ^= xorMask2
	q.UnknownCount ^= xorMask3
	q.BlockFileEnd ^= xorMask4
	q.OffsetFirstFile ^= xorMask1
	q.Version ^= xorMask1  // 1 2
	q.Unknown2 ^= xorMask2 // 0

	// Determines the alignment block size.
	blockShiftBits := 10
	if (q.Flags & 0x800) > 0 {
		blockShiftBits = 12
	}

	sectionsData := make([]byte, 8*q.FileCount)
	if _, err = f.Read(sectionsData); err != nil {
		return fmt.Errorf("cannot read sections: %w", err)
	}

	var sections []uint64
	if sections, err = DecryptSectionList(q.FileCount, sectionsData, q.Version, false); err != nil {
		return fmt.Errorf("cannot decrypt section list: %w", err)
	}

	for _, section := range sections {
		sectionBlock := section >> 40
		sectionOffset := sectionBlock << blockShiftBits
		//slog.Debug("qar read section", "offset", sectionOffset, "raw", fmt.Sprintf("%x", section))
		if _, err = f.Seek(int64(sectionOffset), io.SeekStart); err != nil {
			return fmt.Errorf("cannot seek to section offset %d: %w", sectionOffset, err)
		}

		e := Entry{}
		if err = e.Read(f, q.Version); err != nil {
			return fmt.Errorf("cannot read entry: %w", err)
		}

		q.Entries = append(q.Entries, e)
		//slog.Debug("================================")
	}

	return nil
}

var xorTable = []uint32{
	0x41441043,
	0x11C22050,
	0xD05608C3,
	0x532C7319,
}

func DecryptSectionList(fileCount uint32, sections []byte, version uint32, encrypt bool) ([]uint64, error) {
	result := make([]uint64, fileCount)
	r := bytes.NewReader(sections)

	if version == 2 {
		xor := uint64(0xA2C18EC3)
		for i := 0; i < len(result); i += 1 {
			offset1 := uint64(i * 8)
			offset2 := offset1 + 4

			index1 := (xor + (offset1 / 5)) % 4
			index2 := (xor + (offset2 / 5)) % 4

			s1 := make([]byte, 4)
			if _, err := r.ReadAt(s1, int64(offset1)); err != nil {
				return nil, fmt.Errorf("cannot read section1 at offset %d: %w", offset1, err)
			}
			section1 := binary.LittleEndian.Uint32(s1)

			s2 := make([]byte, 4)
			if _, err := r.ReadAt(s2, int64(offset2)); err != nil {
				return nil, fmt.Errorf("cannot read section2 at offset %d: %w", offset2, err)
			}
			section2 := binary.LittleEndian.Uint32(s2)

			i1 := section1 ^ xorTable[index1]
			i2 := section2 ^ xorTable[index2]
			result[i] = uint64(i2)<<32 | uint64(i1)

			if encrypt {
				i1 = section1
				i2 = section2
			}

			rotation := (int)(i2/256) % 19
			rotated := uint64((i1 >> rotation) | (i1 << (32 - rotation))) // ROR
			xor ^= rotated
		}
		return result, nil
	}

	for i := 0; i < len(result); i += 1 {
		offset1 := uint64(i * 8)
		offset2 := offset1 + 4

		index1 := (uint64(i) + (offset1 / 5)) % 4
		index2 := (uint64(i) + (offset2 / 5)) % 4

		s1 := make([]byte, 4)
		if _, err := r.ReadAt(s1, int64(offset1)); err != nil {
			return nil, fmt.Errorf("cannot read section1 at offset %d: %w", offset1, err)
		}
		section1 := binary.LittleEndian.Uint32(s1)

		s2 := make([]byte, 4)
		if _, err := r.ReadAt(s2, int64(offset2)); err != nil {
			return nil, fmt.Errorf("cannot read section2 at offset %d: %w", offset2, err)
		}
		section2 := binary.LittleEndian.Uint32(s2)

		section1 ^= xorTable[index1]
		section2 ^= xorTable[index2]
		result[i] = uint64(section2)<<32 | uint64(section1)
	}

	return result, nil
}

// ExtractTo qar path to writer. Writer must be closed by user.
func (q *Qar) ExtractTo(path string, hash uint64, writer io.Writer) (int, error) {
	//hash := hashing.HashFileNameWithExtension(path)
	var entry *Entry = nil

	var hashFromName uint64
	var err error
	if filepath.Base(path) == path {
		hash2 := strings.TrimSuffix(path, filepath.Ext(path))
		hashFromName, err = strconv.ParseUint(hash2, 16, 64)
		if err != nil {
			return 0, fmt.Errorf("hashFromName: %w", err)
		}

		hashFromName = hashing.JustAddExtension(hashFromName, filepath.Ext(path))
		//slog.Debug("hashfromname", "value", fmt.Sprintf("0x%x", hashFromName), "ext", filepath.Ext(path)[1:])
	}

	for _, e := range q.Entries {
		//slog.Info("q",
		//	"ph", fmt.Sprintf("%x", e.Header.PathHash),
		//	"hash", fmt.Sprintf("%x", hash),
		//	"hashFromName", fmt.Sprintf("%x", hashFromName))
		if e.Header.PathHash != hash {
			if e.Header.PathHash != hashFromName {
				continue
			}
		}

		entry = &e
		break
	}

	if entry == nil {
		return 0, fmt.Errorf("entry not found, path: %s", path)
	}

	if err = entry.ReadData(q.handle); err != nil {
		return 0, fmt.Errorf("qar entry read data: %w", err)
	}

	n, err := writer.Write(entry.Data)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (q *Qar) Extract(path string, hash uint64, outDir string) (int, error) {
	datDirName := outDir
	workdir := ""
	if outDir == "" {
		workdir = filepath.Dir(q.FilePath)
		datDirName = strings.TrimSuffix(filepath.Base(q.FilePath), ".dat") + "_dat"
	}
	outp := filepath.Dir(path)
	outDir = filepath.Join(workdir, datDirName, outp)
	if err := os.MkdirAll(outDir, os.ModePerm); err != nil {
		return 0, fmt.Errorf("outdir %s: %w", outDir, err)
	}

	outfilePath := filepath.Join(outDir, filepath.Base(path))
	outFile, err := os.OpenFile(outfilePath, os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return 0, fmt.Errorf("extract open output file: %w", err)
	}
	defer outFile.Close()

	n := 0
	if n, err = q.ExtractTo(path, hash, outFile); err != nil {
		return 0, fmt.Errorf("extract: %w", err)
	}

	return n, nil
}

func (q *Qar) SaveDefinition(writer io.Writer) error {
	o, err := json.MarshalIndent(q, "", "  ")
	if err != nil {
		return fmt.Errorf("save definition: %w", err)
	}

	if _, err = writer.Write(o); err != nil {
		return fmt.Errorf("save definition: %w", err)
	}

	return nil
}

func (q *Qar) Write(file io.ReadWriteSeeker, baseDir string) error {
	shift := 10 // 1024
	if q.Flags&0x800 > 0 {
		shift = 12 // 4096
	}

	alignment := 1 << shift

	_, err := file.Seek(int64(HeaderSize+blockSize*len(q.Entries)), io.SeekStart)
	if err != nil {
		return fmt.Errorf("cannot seek to write: %w", err)
	}

	var dataOffset int64
	if dataOffset, err = util.AlignWrite(file, int64(alignment)); err != nil {
		return fmt.Errorf("align write fail: %w", err)
	}

	q.OffsetFirstFile = uint32(dataOffset)
	//slog.Debug("write qar", "offsetFirstFile", q.OffsetFirstFile)

	sections := make([]uint64, len(q.Entries))
	for i, e := range q.Entries {
		slog.Info("writing",
			"entry", e.Header.FilePath,
			"encrypted", fmt.Sprintf("%x", e.DataHeader.EncryptionMagic),
			"key", fmt.Sprintf("%x", e.DataHeader.Key),
		)
		e.Header.Version = q.Version

		if e.Header.NameHashForPacking == 0 {
			e.Header.PathHash = hashing.HashFileNameWithExtension(e.Header.FilePath)
		} else {
			e.Header.PathHash = e.Header.NameHashForPacking
		}

		//slog.Info("v", "ph", fmt.Sprintf("%x", e.Header.PathHash), "p", e.Header.FilePath)
		var pos int64
		pos, err = file.Seek(0, io.SeekCurrent)
		if err != nil {
			return fmt.Errorf("seek entry: %w", err)
		}

		//slog.Debug("qar write section", "pos", pos, "pathHash", fmt.Sprintf("%x", e.Header.PathHash), "filePath", e.Header.FilePath)
		section := uint64((pos>>shift)<<40) | (e.Header.PathHash&0xFF)<<32 | e.Header.PathHash>>32&0xFFFFFFFFFF
		//slog.Debug("qar write section", "value", fmt.Sprintf("%x", section))
		sections[i] = section
		data := []byte{}
		if len(e.Data) == 0 {
			pl := ""
			if pl, err = filepath.Localize(strings.TrimPrefix(e.Header.FilePath, "/")); err != nil {
				return fmt.Errorf("cannot localize entry path %s: %w", pl, err)
			}

			p := filepath.Join(baseDir, e.Header.FilePath)
			if e.Data, err = os.ReadFile(p); err != nil {
				return fmt.Errorf("read entry data: %w", err)
			}
		}
		if data, err = e.Write(); err != nil {
			return fmt.Errorf("write entry to array %s: %w", e.Header.FilePath, err)
		}
		if _, err = file.Write(data); err != nil {
			return fmt.Errorf("write entry %s to file: %w", e.Header.FilePath, err)
		}

		if _, err = util.AlignWrite(file, int64(alignment)); err != nil {
			return fmt.Errorf("entry align fail: %w", err)
		}
	}

	var endPos int64
	if endPos, err = file.Seek(0, io.SeekCurrent); err != nil {
		return fmt.Errorf("endpos seek: %w", err)
	}

	q.BlockFileEnd = uint32(endPos >> shift)
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("seek start: %w", err)
	}

	q.FileCount = uint32(len(q.Entries))

	// error handling? nah
	file.Write(magic[:])
	binary.Write(file, binary.LittleEndian, q.Flags^xorTable[0])
	binary.Write(file, binary.LittleEndian, q.FileCount^xorTable[1])
	binary.Write(file, binary.LittleEndian, q.UnknownCount^xorTable[2]) // not used in GzsTool
	binary.Write(file, binary.LittleEndian, q.BlockFileEnd^xorTable[3])
	binary.Write(file, binary.LittleEndian, q.OffsetFirstFile^xorTable[0])
	binary.Write(file, binary.LittleEndian, q.Version^xorTable[0])
	binary.Write(file, binary.LittleEndian, 0^xorTable[1])

	s, err := q.EncryptSections(sections)
	if err != nil {
		return fmt.Errorf("prepare encrypt sections: %w", err)
	}

	//o, _ := file.Seek(0, io.SeekCurrent)
	//slog.Debug("writing section info at", "offset", o, "len", len(s), "data", fmt.Sprintf("% x", s))
	binary.Write(file, binary.LittleEndian, s)

	//slog.Debug("qar write end\n========")

	return nil
}

//            //long unknownTableOffset = output.Position
//            //output.Skip(16 * UnknownEntries.Count)
//            long dataOffset = output.Position
//            ulong[] sections = new ulong[Entries.Count]
//            for (int i = 0; i < Entries.Count; i++)
//            {
//                QarEntry entry = Entries[i]
//                entry.CalculateHash()
//                ulong section = (ulong) (output.Position >> shift) << 40
//                                | (entry.PathHash & 0xFF) << 32
//                                | entry.PathHash >> 32 & 0xFFFFFFFFFF
//                sections[i] = section
//                entry.Write(output, inputDirectory)
//                output.AlignWrite(alignment, 0x00)
//            }
//            long endPosition = output.Position
//            uint endPositionHead = (uint) (endPosition >> shift)
//
//            output.Position = headerPosition
//            writer.Write(QarMagicNumber); // SQAR
//            writer.Write(Flags ^ xorMask1)
//            writer.Write((uint)Entries.Count ^ xorMask2)
//            writer.Write(xorMask3); // unknown count (not saved in the xml and output directory)
//            writer.Write(endPositionHead ^ xorMask4)
//            writer.Write((uint)dataOffset ^ xorMask1)
//            writer.Write(Version ^ xorMask1)
//            writer.Write(0 ^ xorMask2)
//
//            output.Position = tableOffset
//            byte[] encryptedSectionsData = EncryptSections(sections)
//            writer.Write(encryptedSectionsData)
//
//            output.Position = endPosition
//        }

func (q *Qar) EncryptSections(sections []uint64) ([]byte, error) {
	out := make([]byte, len(sections)*blockSize)
	for i, v := range sections {
		binary.LittleEndian.PutUint64(out[i*8:], v)
	}

	encSections, err := DecryptSectionList(q.FileCount, out, q.Version, true)
	if err != nil {
		return nil, fmt.Errorf("encrypt sections: %w", err)
	}

	for i, v := range encSections {
		binary.LittleEndian.PutUint64(out[i*8:], v)
		//slog.Debug("put", "x", fmt.Sprintf("%x", out[i*8:i*8+8]))
	}

	return out, nil
}
