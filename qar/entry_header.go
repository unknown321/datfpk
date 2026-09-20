package qar

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/unknown321/hashing"
)

const HeaderSize = 32
const HeaderSizeConsole = 16

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

func (e *EntryHeader) ReadConsole(r io.ReadSeeker, byteOrder binary.ByteOrder) error {
	var err error

	if err = binary.Read(r, byteOrder, &e.PathHash); err != nil {
		return fmt.Errorf("cannot read path hash: %w", err)
	}

	var dataOffset uint32
	if err = binary.Read(r, byteOrder, &dataOffset); err != nil {
		return fmt.Errorf("cannot read offset: %w", err)
	}
	e.DataOffset = int64(dataOffset * 16)

	if err = binary.Read(r, byteOrder, &e.UncompressedSize); err != nil {
		return fmt.Errorf("cannot read uncompressed size: %w", err)
	}

	return nil
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
	var res []byte
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

func (e *EntryHeader) BytesConsole() ([]byte, error) {
	res := make([]byte, 8+4+4)

	if e.NameHashForPacking == 0 {
		e.PathHash = hashing.HashFileNameWithExtension(e.FilePath)
	} else {
		e.PathHash = e.NameHashForPacking
	}

	binary.LittleEndian.PutUint64(res[0:], e.PathHash)
	binary.LittleEndian.PutUint32(res[8:], uint32(e.DataOffset/16))
	binary.LittleEndian.PutUint32(res[12:], e.UncompressedSize)

	return res, nil
}
