package qar

import (
	"bytes"
	"compress/zlib"
	"crypto/md5"
	"datfpk/crypto"
	"datfpk/hashing"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
)

type EntryHeader struct {
	// read from binary
	PathHash         uint64 `json:"-"` // cityHash(FilePath), 8
	UncompressedSize uint32 `json:"-"` // 4
	CompressedSize   uint32 `json:"-"` // 4
	Md5Sum           Md5Sum `json:"-"` // md5sum(compressed?(encrypted?(data))), 16

	// derived properties
	FilePath   string `json:"filePath"`
	DataOffset int64  `json:"-"` // offset from the beginning of the file
	Compressed bool   `json:"compressed,omitempty"`
	Version    uint32 `json:"-"`
	MetaFlag   bool   `json:"meta_flag,omitempty"`

	NameHashForPacking uint64 `json:"hash,omitempty"` // used for packing when name is not resolved
}

func (e *EntryHeader) Read(reader io.ReadSeeker, version uint32) error {
	e.Version = version

	hashLow, err := ReadUint32(reader)
	if err != nil {
		return err
	}
	hashLow ^= xorMask1

	hashHigh, err := ReadUint32(reader)
	if err != nil {
		return err
	}
	hashHigh ^= xorMask1

	e.PathHash = uint64(hashHigh)<<32 | uint64(hashLow)
	//o, _ := reader.Seek(0, io.SeekCurrent)
	//slog.Debug("entry header read", "pathHash", fmt.Sprintf("%x", e.PathHash), "offset", o)

	size1, err := ReadUint32(reader)
	if err != nil {
		return err
	}
	size1 ^= xorMask2

	size2, err := ReadUint32(reader)
	if err != nil {
		return err
	}
	size2 ^= xorMask3

	e.CompressedSize = size1
	e.UncompressedSize = size2
	//slog.Debug("entry header read", "compressed", size1, "uncompressed", size2)

	//if e.Version == 2 {
	//	e.CompressedSize = size1
	//	e.UncompressedSize = size2
	//}
	if e.CompressedSize != e.UncompressedSize {
		e.Compressed = true
	}

	if _, err = reader.Read(e.Md5Sum[:]); err != nil {
		return fmt.Errorf("read md5: %w", err)
	}

	//slog.Debug("entry header read md5", "md5 encoded", fmt.Sprintf("%x", e.Md5Sum))
	if e.Md5Sum, err = Md5Decode(e.Md5Sum); err != nil {
		return fmt.Errorf("md5 decode: %w", err)
	}

	// reading done
	if e.DataOffset, err = reader.Seek(0, io.SeekCurrent); err != nil {
		return fmt.Errorf("cannot get offset: %w", err)
	}

	e.MetaFlag = (e.PathHash & hashing.MetaFlag) > 0

	//slog.Debug("entry header read done", "md5", fmt.Sprintf("%x", e.Md5Sum), "dataOffset", e.DataOffset, "metaFlag", e.MetaFlag)

	return nil
}

func (e *EntryHeader) Bytes() ([]byte, error) {
	res := []byte{}
	writer := bytes.NewBuffer(res)
	var err error

	if e.NameHashForPacking == 0 {
		e.PathHash = hashing.HashFileNameWithExtension(e.FilePath)
	} else {
		e.PathHash = e.NameHashForPacking
	}

	//slog.Debug("write header", "pathHash", fmt.Sprintf("%x", e.PathHash), "filePath", e.FilePath)
	if err = binary.Write(writer, binary.LittleEndian, e.PathHash^xorMask1Long); err != nil {
		return nil, err
	}

	//size1 := e.UncompressedSize
	//size2 := e.CompressedSize
	//if e.Version == 2 {
	//	size1 = e.CompressedSize
	//	size2 = e.UncompressedSize
	//}
	if err = binary.Write(writer, binary.LittleEndian, e.CompressedSize^xorMask2); err != nil {
		return nil, err
	}
	if err = binary.Write(writer, binary.LittleEndian, e.UncompressedSize^xorMask3); err != nil {
		return nil, err
	}
	//slog.Debug("write header", "compressed", e.CompressedSize, "uncompressed", e.UncompressedSize)

	m, err := Md5Decode(e.Md5Sum)
	if err != nil {
		return nil, err
	}
	if err = binary.Write(writer, binary.LittleEndian, m); err != nil {
		return nil, err
	}

	e.DataOffset = HeaderSize
	//slog.Debug("write header", "dataOffset", e.DataOffset)

	return writer.Bytes(), nil
}

type DataHeader struct {
	EncryptionMagic  uint32 `json:"encryption,omitempty"`
	Key              uint32 `json:"key,omitempty"`
	CompressedSize   uint32
	UncompressedSize uint32
}

func (dh *DataHeader) Parse(data []byte) {
	dh.EncryptionMagic = binary.LittleEndian.Uint32(data)
	//slog.Debug("data header", "magic", fmt.Sprintf("%x", dh.EncryptionMagic))
	if dh.EncryptionMagic == crypto.Magic1 || dh.EncryptionMagic == crypto.Magic2 {
		//slog.Debug("encrypted")
		dh.Key = binary.LittleEndian.Uint32(data[4:])
	} else {
		dh.EncryptionMagic = 0
	}
}

func (dh *DataHeader) Bytes() []byte {
	s := crypto.GetHeaderSize(dh.EncryptionMagic)
	if s == 0 {
		return nil
	}
	out := make([]byte, s)
	binary.LittleEndian.PutUint32(out[0:], dh.EncryptionMagic)
	binary.LittleEndian.PutUint32(out[4:], dh.Key)
	if s > 8 {
		binary.LittleEndian.PutUint32(out[8:], dh.UncompressedSize)
		binary.LittleEndian.PutUint32(out[12:], dh.CompressedSize)
	}
	return out
}

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

const xorMask1 = uint32(0x41441043)
const xorMask2 = uint32(0x11C22050)
const xorMask3 = uint32(0xD05608C3)
const xorMask4 = uint32(0x532C7319)
const xorMask1Long = uint64(0x4144104341441043)

const HeaderSize = 32

func ReadUint32(reader io.Reader) (uint32, error) {
	i32 := make([]byte, 4)
	var err error
	if _, err = reader.Read(i32); err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint32(i32), nil
}

func (e *Entry) Read(reader io.ReadSeeker, version uint32) error {
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

	e.DataHeader.Parse(dh)
	//slog.Info("header", "v", fmt.Sprintf("%+v", e.Header))
	//slog.Info("dataHeader", "v", fmt.Sprintf("%x, %+v", dh, e.DataHeader))
	//slog.Info("Entry",
	//	"hash", fmt.Sprintf("%x", e.Header.PathHash),
	//	"uncompSize", e.Header.UncompressedSize,
	//	"compSize", e.Header.CompressedSize,
	//	"key", fmt.Sprintf("%x", e.DataHeader.Key),
	//	"magic", fmt.Sprintf("%x", e.DataHeader.EncryptionMagic))

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
	b := []byte{}
	if b, err = d1.Read(bytes.NewReader(mdata), len(mdata)); err != nil {
		return nil, fmt.Errorf("write decrypt1stream: %w", err)
	}

	res := append(header, b...)
	//slog.Debug("write done", "res", len(res), "header", len(header), "dataH", len(dataHeader), "data", len(b))
	//slog.Debug("header", "v", fmt.Sprintf("%x, %+v", header, e.Header))
	//slog.Debug("dataHeader", "v", fmt.Sprintf("%x, %+v", dataHeader, e.DataHeader))

	return res, nil
}

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

	//slog.Debug("read data",
	//	"size", size,
	//	"offset", e.Header.DataOffset,
	//	"uncompSize", e.Header.UncompressedSize,
	//	"compSize", e.Header.CompressedSize,
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

//func (e *Entry) Write2() ([]byte, error) {
//	e.Header.CompressedSize = uint32(len(e.Data))
//	e.Header.UncompressedSize = uint32(len(e.Data))
//
//	e.DataHeader.EncryptionMagic = crypto.Magic2
//
//	var err error
//	entryData := []byte{}
//	if e.DataHeader.EncryptionMagic > 0 {
//		slog.Info("encrypting")
//		e.DataHeader.CompressedSize = e.Header.CompressedSize
//		e.DataHeader.UncompressedSize = e.Header.UncompressedSize
//		e.Header.CompressedSize += uint32(crypto.GetHeaderSize(e.DataHeader.EncryptionMagic))
//		e.Header.UncompressedSize += uint32(crypto.GetHeaderSize(e.DataHeader.EncryptionMagic))
//		d2 := Decrypt2Stream{}
//		d2.Init(e.DataHeader.Key)
//		entryData, err = d2.Read(bytes.NewBuffer(e.Data), 0, len(e.Data))
//		if err != nil {
//			return nil, fmt.Errorf("write decrypt2 data: %w", err)
//		}
//	}
//
//	dataHeader := e.DataHeader.Bytes()
//
//	mdata := append(dataHeader, entryData...)
//	md5sum := md5.Sum(mdata)
//	e.Header.Md5Sum = md5sum
//	slog.Info("md5", "v", fmt.Sprintf("%x", md5sum), "from", fmt.Sprintf("%x", mdata))
//	header, err := e.Header.Bytes()
//
//	d1 := Decrypt1Stream{}
//	d1.Init(e.Header.Md5Sum, e.Header.PathHash, e.Header.Version, len(mdata))
//	b := []byte{}
//	if b, err = d1.Read(bytes.NewReader(mdata), len(mdata)); err != nil {
//		return nil, fmt.Errorf("decrypt1stream: %w", err)
//	}
//
//	//res := append(header, dataHeaderEncrypted...)
//	res := append(header, b...)
//	slog.Info("write done", "res", len(res), "header", len(header), "dataH", len(dataHeader), "data", len(b))
//	slog.Info("header", "v", fmt.Sprintf("%x, %+v", header, e.Header))
//	slog.Info("dataHeader", "v", fmt.Sprintf("%x, %+v", dataHeader, e.DataHeader))
//
//	return res, err
//}
