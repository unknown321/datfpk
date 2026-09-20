package qar

import (
	"encoding/binary"
	"fmt"
	"log/slog"

	"github.com/unknown321/datfpk/crypto"
)

type DataHeader struct {
	EncryptionMagic  uint32 `json:"encryption,omitempty"`
	Key              uint32 `json:"key,omitempty"`
	CompressedSize   uint32
	UncompressedSize uint32
	Version          uint32 `json:"version,omitempty"`
	byteOrder        binary.ByteOrder
}

func (dh *DataHeader) Read(data []byte, byteOrder binary.ByteOrder) {
	dh.byteOrder = byteOrder
	dh.EncryptionMagic = dh.byteOrder.Uint32(data)
	slog.Debug("data header", "magic", fmt.Sprintf("%x", dh.EncryptionMagic))
	switch dh.EncryptionMagic {
	case crypto.Magic1:
		dh.Version = 1
	case crypto.Magic2:
		dh.Version = 2
	default:
		dh.Version = 0
	}

	if dh.Version != 0 {
		dh.Key = dh.byteOrder.Uint32(data[4:])
	} else {
		dh.EncryptionMagic = 0
	}
}

func (dh *DataHeader) ReadConsole(data []byte, byteOrder binary.ByteOrder) error {
	dh.byteOrder = byteOrder

	dh.EncryptionMagic = dh.byteOrder.Uint32(data)
	dh.Key = dh.byteOrder.Uint32(data[4:])
	dh.UncompressedSize = dh.byteOrder.Uint32(data[8:])

	return nil
}

// Bytes writes data header bytes into byte array
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

func (dh *DataHeader) BytesConsole() []byte {
	s := crypto.GetHeaderSize(dh.EncryptionMagic)
	if s == 0 {
		return nil
	}

	out := make([]byte, s)

	dh.byteOrder.PutUint32(out[0:], dh.EncryptionMagic)
	dh.byteOrder.PutUint32(out[4:], dh.Key)
	dh.byteOrder.PutUint32(out[8:], dh.UncompressedSize)
	dh.byteOrder.PutUint32(out[12:], dh.CompressedSize)

	return out
}
