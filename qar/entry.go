package qar

import (
	"bytes"
	"compress/zlib"
	"crypto/md5"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"

	"github.com/unknown321/datfpk/crypto"
)

var ConsoleMagic = [4]byte{0xe6, 0xef, 0xf8, 0xe3} // xbox?
//var anotherConsoleMagic = [4]byte{0xe6, 0xef, 0xf8, 0xa0} // not xbox? offset is -8

const xorMask1 = uint32(0x41441043)
const xorMask2 = uint32(0x11C22050)
const xorMask3 = uint32(0xD05608C3)
const xorMask4 = uint32(0x532C7319)
const xorMask1Long = uint64(0x4144104341441043)

const consoleKeyMul = 0x2e90edd
const consoleKeyXor = 0x6576

type Entry struct {
	Header     EntryHeader
	DataHeader DataHeader
	Data       []byte `json:"-"`
}

func (e *Entry) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		FilePath   string `json:"filePath"`
		Compressed bool   `json:"compressed,omitempty"`
		MetaFlag   bool   `json:"metaFlag,omitempty"`
		Encryption uint32 `json:"encryption,omitempty"`
		Key        uint32 `json:"key,omitempty"`
		Hash       uint64 `json:"hash,omitempty"` // used only for files without resolved names
	}{
		FilePath:   e.Header.FilePath,
		Compressed: e.Header.Compressed,
		MetaFlag:   e.Header.MetaFlag,
		Encryption: e.DataHeader.EncryptionMagic,
		Key:        e.DataHeader.Key,
		Hash:       e.Header.NameHashForPacking,
	})
}

func (e *Entry) UnmarshalJSON(i []byte) error {
	if err := json.Unmarshal(i, &e.Header); err != nil {
		return err
	}

	if err := json.Unmarshal(i, &e.DataHeader); err != nil {
		return err
	}

	return nil
}

func ReadUint32(reader io.Reader) (uint32, error) {
	i32 := make([]byte, 4)
	var err error
	if _, err = reader.Read(i32); err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint32(i32), nil
}

// Read reads entry header and data header only
func (e *Entry) Read(reader io.ReadSeeker, version uint32, byteOrder binary.ByteOrder) error {
	var err error

	if err = e.Header.Read(reader, version); err != nil {
		return fmt.Errorf("read header: %w", err)
	}

	d1 := Decrypt1Stream{}
	d1.Init(e.Header.Md5Sum, e.Header.PathHash, version, 8)
	dh, err := d1.Read(reader, 8)
	if err != nil {
		return fmt.Errorf("data header: %w", err)
	}

	e.DataHeader.Read(dh, byteOrder)

	return nil
}

func (e *Entry) Write() ([]byte, error) {
	e.Header.CompressedSize = uint32(len(e.Data))
	e.Header.UncompressedSize = uint32(len(e.Data))

	entryData := e.Data

	if e.Header.Compressed {
		//slog.Debug("write compressed", "uncompressedSize", e.Header.UncompressedSize, "data", fmt.Sprintf("% x", e.Data))
		b := new(bytes.Buffer)
		z, err := zlib.NewWriterLevel(b, zlib.BestCompression)
		if err != nil {
			return nil, fmt.Errorf("compress create: %w", err)
		}
		if _, err = z.Write(entryData); err != nil {
			_ = z.Close()
			return nil, fmt.Errorf("compress: %w", err)
		}
		_ = z.Close()

		e.Header.UncompressedSize = uint32(len(e.Data))
		entryData = b.Bytes()
		e.Header.CompressedSize = uint32(len(b.Bytes()))
		//slog.Debug("data", "in", fmt.Sprintf("%s", e.Data), "out", fmt.Sprintf("% x", b.Bytes()), "compSize", len(b.Bytes()))
	}

	if e.DataHeader.Key > 0 {
		e.DataHeader.EncryptionMagic = crypto.Magic2

		e.DataHeader.CompressedSize = e.Header.CompressedSize
		e.DataHeader.UncompressedSize = e.Header.UncompressedSize
		hs := crypto.GetHeaderSize(e.DataHeader.EncryptionMagic)
		e.Header.UncompressedSize += uint32(hs)
		e.Header.CompressedSize += uint32(hs)
		d2 := Decrypt2Stream{}
		d2.Init(e.DataHeader.Key)
		var err error
		entryData, err = d2.Read(bytes.NewBuffer(entryData), len(entryData))
		if err != nil {
			return nil, fmt.Errorf("write encrypt d2: %w", err)
		}
	}

	dataHeader := e.DataHeader.Bytes()
	mdata := append(dataHeader, entryData...)
	md5sum := md5.Sum(mdata)
	e.Header.Md5Sum = md5sum
	//slog.Debug("md5", "v", fmt.Sprintf("%x", md5sum), "from", fmt.Sprintf("%x", mdata))
	header, err := e.Header.Bytes()

	//slog.Debug("write entry",
	//	"filePath", e.Header.FilePath,
	//	"pathHash", fmt.Sprintf("%x", e.Header.PathHash),
	//	"key", fmt.Sprintf("%x", e.DataHeader.Key),
	//	"magic", fmt.Sprintf("%x", e.DataHeader.EncryptionMagic),
	//)

	d1 := Decrypt1Stream{}
	d1.Init(e.Header.Md5Sum, e.Header.PathHash, e.Header.Version, len(mdata))
	var b []byte
	if b, err = d1.Read(bytes.NewReader(mdata), len(mdata)); err != nil {
		return nil, fmt.Errorf("write decrypt1stream: %w", err)
	}

	res := append(header, b...)
	//slog.Debug("write done", "res", len(res), "header", len(header), "dataH", len(dataHeader), "data", len(b))
	//slog.Debug("header", "v", fmt.Sprintf("%x, %+v", header, e.Header))
	//slog.Debug("dataHeader", "v", fmt.Sprintf("%x, %+v", dataHeader, e.DataHeader))

	return res, nil
}

// ReadDataConsole reads data header (if encrypted) + entry data
func (e *Entry) ReadDataConsole(f io.ReadSeeker) error {
	var err error
	var data = make([]byte, e.Header.UncompressedSize)
	n := 0
	if _, err = f.Seek(e.Header.DataOffset, io.SeekStart); err != nil {
		return err
	}

	if n, err = f.Read(data); err != nil {
		return err
	}

	if n != int(e.Header.UncompressedSize) {
		return fmt.Errorf("unexpected read size from file, want %d, got %d", e.Header.UncompressedSize, n)
	}

	// not encrypted
	if bytes.Compare(data[0:4], ConsoleMagic[:]) != 0 {
		e.Data = data

		return nil
	}

	// encrypted
	{
		if err = e.DataHeader.ReadConsole(data, binary.LittleEndian); err != nil {
			return err
		}

		dataHeaderSize := crypto.GetHeaderSize(e.DataHeader.EncryptionMagic)

		e.Data = make([]byte, e.Header.UncompressedSize-uint32(dataHeaderSize))
		key := int32(e.DataHeader.Key)
		uncompressedSize := int(e.DataHeader.UncompressedSize)

		var key2 int32
		key2 = (key << 0x5) - key

		keyStart := (key^consoleKeyXor)<<0x10 | key

		var val int32
		var offset = 0
		var fullBlocks = uncompressedSize / 4
		var remainder = uncompressedSize % 4
		for offset = 0; offset < fullBlocks; offset++ {
			dataBlock := binary.LittleEndian.Uint32(data[dataHeaderSize+offset*4:])
			val = keyStart ^ int32(dataBlock)
			binary.LittleEndian.PutUint32(e.Data[offset*4:], uint32(val))
			keyStart = keyStart*consoleKeyMul + key2*0x8 + key2 - key
		}

		for nn, b := range data[dataHeaderSize+offset*4 : dataHeaderSize+offset*4+remainder] {
			e.Data[offset*4+nn] = b
		}
	}

	return nil
}

// WriteConsole writes data header (if encrypted) + entry data
func (e *Entry) WriteConsole() ([]byte, error) {
	e.DataHeader.byteOrder = binary.LittleEndian
	e.DataHeader.UncompressedSize = uint32(len(e.Data))
	e.DataHeader.CompressedSize = e.DataHeader.UncompressedSize

	if e.DataHeader.Key == 0 {
		return e.Data, nil
	}

	e.DataHeader.EncryptionMagic = crypto.Magic2

	dataHeaderSize := crypto.GetHeaderSize(e.DataHeader.EncryptionMagic)
	out := make([]byte, dataHeaderSize+len(e.Data))

	copy(out, e.DataHeader.BytesConsole())

	key := int32(e.DataHeader.Key)
	copy(out[dataHeaderSize:], e.Data)

	key2 := (key << 0x5) - key
	keyStart := (key^consoleKeyXor)<<0x10 | key

	var val int32
	var fullBlocks = len(e.Data) / 4
	for offset := range fullBlocks {
		dataBlock := binary.LittleEndian.Uint32(out[dataHeaderSize+offset*4:])
		val = keyStart ^ int32(dataBlock)
		binary.LittleEndian.PutUint32(out[dataHeaderSize+offset*4:], uint32(val))
		keyStart = keyStart*consoleKeyMul + key2*0x8 + key2 - key
	}

	e.Header.UncompressedSize = uint32(len(out))
	e.Header.CompressedSize = e.Header.UncompressedSize

	return out, nil
}

// ReadData reads entry data for pc (separate from headers)
func (e *Entry) ReadData(reader io.ReadSeeker) error {
	_, err := reader.Seek(e.Header.DataOffset, io.SeekStart)
	//slog.Debug("read data", "off", e.Header.DataOffset)
	if err != nil {
		return err
	}
	size := int(e.Header.UncompressedSize)
	if e.Header.CompressedSize > e.Header.UncompressedSize {
		size = int(e.Header.CompressedSize)
	}

	//slog.Info("read data",
	//	"size", size,
	//	"offset", e.Header.DataOffset,
	//	"uncompSize", e.Header.UncompressedSize,
	//	"compSize", e.Header.CompressedSize,
	//	"version", e.Header.Version,
	//	"sum", e.Header.Md5Sum,
	//	"pathHash", e.Header.PathHash,
	//)

	d1 := Decrypt1Stream{}
	d1.Init(e.Header.Md5Sum, e.Header.PathHash, e.Header.Version, size)
	//o, _ := reader.Seek(0, io.SeekCurrent)
	//slog.Debug("reading data from", "offset", o)
	e.Data, err = d1.Read(reader, size)
	if err != nil {
		return err
	}

	//md5sum := md5.Sum(e.Data)
	//if bytes.Compare(e.Header.Md5Sum[:], md5sum[:]) != 0 {
	//	return fmt.Errorf("md5 mismatch, got %x, want %x", md5sum, e.Header.Md5Sum)
	//}
	//slog.Debug("md5 from", "val", fmt.Sprintf("%x", md5sum), "data", fmt.Sprintf("%x", e.Data))

	if e.DataHeader.EncryptionMagic > 0 {
		headerSize := crypto.GetHeaderSize(e.DataHeader.EncryptionMagic)
		if _, err = reader.Seek(int64(headerSize), io.SeekCurrent); err != nil {
			return fmt.Errorf("seek failed: %w", err)
		}
		size -= headerSize
		//slog.Debug("data header", "size", headerSize, "new size", size)

		d2 := Decrypt2Stream{}
		d2.Init(e.DataHeader.Key)
		e.Data, err = d2.Read(bytes.NewBuffer(e.Data[headerSize:]), size)
		if err != nil {
			return fmt.Errorf("read decrypt2 data: %w", err)
		}
		//slog.Debug("read data post d2", "len", len(e.Data))
	}

	if e.Header.Compressed {
		//slog.Debug("entry compressed", "z", fmt.Sprintf("% x", e.Data), "c", e.Header.CompressedSize, "len", len(e.Data), "u", e.Header.UncompressedSize)
		var z io.ReadCloser
		r := bytes.NewReader(e.Data)
		z, err = zlib.NewReader(r)

		defer z.Close()

		if err != nil {
			return fmt.Errorf("create zlib reader: %w", err)
		}
		var b bytes.Buffer

		_, err = io.Copy(&b, z)
		if err != nil {
			return fmt.Errorf("decompress zlib: %w", err)
		}

		e.Data = b.Bytes()[0:e.Header.UncompressedSize]
	}

	return nil
}
