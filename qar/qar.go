package qar

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/unknown321/datfpk/util"

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

	handle    io.ReadSeeker
	isConsole bool
}

var PCMagic = [4]byte{0x53, 0x51, 0x41, 0x52} // SQAR

const ID = "qar"
const IDConsole = "qarconsole"

type qjs struct {
	Type    string  `json:"type"`
	Flags   uint32  `json:"flags"`
	Version uint32  `json:"version"`
	Entries []Entry `json:"entries"`
}

func (q *Qar) MarshalJSON() ([]byte, error) {
	id := ID
	// todo possible difference between consoles?
	switch q.isConsole {
	case true:
		id = IDConsole
	}

	return json.Marshal(qjs{
		Type:    id,
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

	switch qq.Type {
	case ID:
		break
	case IDConsole:
		q.isConsole = true
	default:
		return fmt.Errorf("unexpected type %s, want %s or %s", qq.Type, ID, IDConsole)
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

	q.isConsole = true
	if bytes.Compare(q.Magic[:], PCMagic[:]) == 0 {
		q.isConsole = false
	}

	if q.isConsole {
		err = q.readConsole(f)
	} else {
		err = q.readPC(f)
	}

	if err != nil {
		return err
	}

	return nil
}

// readConsole reads qar header and entries headers, no data
func (q *Qar) readConsole(f io.ReadSeeker) error {
	byteOrder := binary.LittleEndian

	var err error

	// qar header
	{
		if _, err = f.Seek(-4, io.SeekEnd); err != nil {
			return err
		}

		if err = binary.Read(f, byteOrder, &q.Version); err != nil {
			return err
		}

		var mysticVar = 0
		if q.Version == 0x24 {
			mysticVar = 0x8b
		}
		if q.Version == 0x14 {
			mysticVar = 0x6b
		}

		if mysticVar == 0 {
			return fmt.Errorf("unsupported version, got %x, want 0x14 or 0x24", q.Version)
		}

		if _, err = f.Seek(-20, io.SeekEnd); err != nil {
			return err
		}

		if err = binary.Read(f, byteOrder, &q.FileCount); err != nil {
			return err
		}

		if _, err = f.Seek(-36, io.SeekEnd); err != nil {
			return err
		}
	}

	q.Entries = make([]Entry, q.FileCount)
	if _, err = q.handle.Seek(-1*HeaderSizeConsole*int64(q.FileCount), io.SeekCurrent); err != nil {
		return fmt.Errorf("cannot seek to the entry table start")
	}

	for i := range q.FileCount {
		if err = q.Entries[i].Header.ReadConsole(q.handle, binary.LittleEndian); err != nil {
			return err
		}
	}

	return nil
}

// readPC reads qar header and entries headers
func (q *Qar) readPC(f io.ReadSeeker) error {
	byteOrder := binary.LittleEndian
	var err error

	if err = binary.Read(f, byteOrder, &q.Flags); err != nil {
		return fmt.Errorf("cannot read flags: %w", err)
	}

	if err = binary.Read(f, byteOrder, &q.FileCount); err != nil {
		return fmt.Errorf("cannot read file count: %w", err)
	}

	if err = binary.Read(f, byteOrder, &q.UnknownCount); err != nil {
		return fmt.Errorf("cannot read unknown count: %w", err)
	}

	if err = binary.Read(f, byteOrder, &q.BlockFileEnd); err != nil {
		return fmt.Errorf("cannot read block file end: %w", err)
	}

	if err = binary.Read(f, byteOrder, &q.OffsetFirstFile); err != nil {
		return fmt.Errorf("cannot read offset first file: %w", err)
	}

	if err = binary.Read(f, byteOrder, &q.Version); err != nil {
		return fmt.Errorf("cannot read version: %w", err)
	}

	if err = binary.Read(f, byteOrder, &q.Unknown2); err != nil {
		return fmt.Errorf("cannot read unknown2: %w", err)
	}

	q.Flags ^= xorMask1
	q.FileCount ^= xorMask2
	q.UnknownCount ^= xorMask3
	q.BlockFileEnd ^= xorMask4
	q.OffsetFirstFile ^= xorMask1
	q.Version ^= xorMask1  // 1 2
	q.Unknown2 ^= xorMask2 // 0

	sectionsData := make([]byte, 8*q.FileCount)
	if _, err = f.Read(sectionsData); err != nil {
		return fmt.Errorf("cannot read sections: %w", err)
	}

	var sections []uint64
	if sections, err = DecryptSectionList(q.FileCount, sectionsData, q.Version, false); err != nil {
		return fmt.Errorf("cannot decrypt section list: %w", err)
	}

	// Determines the alignment block size.
	blockShiftBits := 10
	if (q.Flags & 0x800) > 0 {
		blockShiftBits = 12
	}

	for _, section := range sections {
		sectionBlock := section >> 40
		sectionOffset := sectionBlock << blockShiftBits
		//slog.Debug("qar read section", "offset", sectionOffset, "raw", fmt.Sprintf("%x", section))
		if _, err = f.Seek(int64(sectionOffset), io.SeekStart); err != nil {
			return fmt.Errorf("cannot seek to section offset %d: %w", sectionOffset, err)
		}

		e := Entry{}
		if err = e.Read(f, q.Version, byteOrder); err != nil {
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
			hash2 = strings.TrimSuffix(hash2, filepath.Ext(hash2)) // entry=38dd243657e7f.2.ftexs
			if hashFromName, err = strconv.ParseUint(hash2, 16, 64); err != nil {
				return 0, fmt.Errorf("hashFromName: %w", err)
			}
		}

		hashFromName = hashing.JustAddExtension(hashFromName, filepath.Ext(path))
		//slog.Debug("hashfromname", "value", fmt.Sprintf("0x%x", hashFromName), "ext", filepath.Ext(path)[1:])
	}

	// O(n), but we keep order
	for n, _ := range q.Entries {
		//slog.Info("q",
		//	"ph", fmt.Sprintf("%x", e.Header.PathHash),
		//	"hash", fmt.Sprintf("%x", hash),
		//	"hashFromName", fmt.Sprintf("%x", hashFromName))
		if q.Entries[n].Header.PathHash != hash {
			if q.Entries[n].Header.PathHash != hashFromName {
				continue
			}
		}

		entry = &q.Entries[n]
		break
	}

	if entry == nil {
		return 0, fmt.Errorf("entry not found, path: %s", path)
	}

	if q.isConsole {
		if err = entry.ReadDataConsole(q.handle); err != nil {
			return 0, fmt.Errorf("qar entry read data: %w", err)
		}
	} else {
		if err = entry.ReadData(q.handle); err != nil {
			return 0, fmt.Errorf("qar entry read data: %w", err)
		}
	}

	n, err := writer.Write(entry.Data)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// Extract extracts entry with <hash> to <outDir>, optionally using <path> as output filename
// example: extract file with hash 0x123456 to /tmp/test with name /assets/test:
//
//	q.Extract("/assets/test", 0x123456, "/tmp/test")
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
		return 0, fmt.Errorf("qar extract: %w", err)
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

func (q *Qar) WriteConsole(file io.ReadWriteSeeker, baseDir string, printLog bool) error {
	var err error
	alignment := int64(16)

	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("write console seek: %w", err)
	}

	dataSize := 0
	q.FileCount = uint32(len(q.Entries))
	for i, e := range q.Entries {
		if q.Entries[i].Header.DataOffset, err = file.Seek(0, io.SeekCurrent); err != nil {
			return fmt.Errorf("write console seek while writing entries: %w", err)
		}

		if printLog {
			slog.Info("writing",
				"entry", e.Header.FilePath,
				"encrypted", fmt.Sprintf("%x", e.DataHeader.EncryptionMagic),
				"key", fmt.Sprintf("%x", e.DataHeader.Key),
			)
		}

		var data []byte
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

		data, err = e.WriteConsole()
		if err != nil {
			return fmt.Errorf("write console entry: %w", err)
		}

		q.Entries[i].Header.UncompressedSize = uint32(len(data))
		if q.Entries[i].Header.Compressed {
			return fmt.Errorf("console compression not implemented")
		}
		dataSize += len(data)

		if _, err = file.Write(data); err != nil {
			return fmt.Errorf("write console entry to file: %w", err)
		}

		if _, err = util.AlignWrite(file, alignment); err != nil {
			return fmt.Errorf("write console alignment: %w", err)
		}
	}

	entryTableOffset, err := file.Seek(0, io.SeekCurrent)
	if err != nil {
		return fmt.Errorf("write console, get entry table offset: %w", err)
	}

	entryTableOffset /= 16

	for _, entry := range q.Entries {
		data, err1 := entry.Header.BytesConsole()
		if err1 != nil {
			return fmt.Errorf("write console entry header: %w", err1)
		}

		if _, err = file.Write(data); err != nil {
			return fmt.Errorf("write console entry header to file: %w", err)
		}
	}

	// header
	{
		binary.Write(file, binary.LittleEndian, uint32(dataSize))
		binary.Write(file, binary.LittleEndian, uint32(0)) // ??
		binary.Write(file, binary.LittleEndian, uint32(0)) // ??
		binary.Write(file, binary.LittleEndian, uint32(0)) // ??
		binary.Write(file, binary.LittleEndian, q.FileCount)
		binary.Write(file, binary.LittleEndian, uint32(0x71610000))       // ??, same for ps3 and xbox360
		binary.Write(file, binary.LittleEndian, uint32(entryTableOffset)) // ??
		binary.Write(file, binary.LittleEndian, uint32(0))                // ??
		binary.Write(file, binary.LittleEndian, q.Version)
	}

	return nil
}

// Write creates new QAR file
func (q *Qar) Write(file io.ReadWriteSeeker, baseDir string, printLog bool) error {
	q.handle = file
	if q.isConsole {
		return q.WriteConsole(file, baseDir, printLog)
	}

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
		if printLog {
			slog.Info("writing",
				"entry", e.Header.FilePath,
				"encrypted", fmt.Sprintf("%x", e.DataHeader.EncryptionMagic),
				"key", fmt.Sprintf("%x", e.DataHeader.Key),
			)
		}
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
	file.Write(PCMagic[:])
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

func (q *Qar) ReadFile(path string) ([]byte, error) {
	ph := hashing.HashFileNameWithExtension(path)
	for _, v := range q.Entries {
		if v.Header.PathHash == ph {
			if err := v.ReadData(q.handle); err != nil {
				return nil, err
			}

			return v.Data, nil
		}
	}

	return nil, fmt.Errorf("file not found")
}
