package fpk

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type Header struct {
	Magic        [10]byte
	FileSize     uint32
	_            [18]byte
	MagicNumber2 uint32
	EntryCount   uint32
	RefCount     uint32
	_            uint32
}

const HeaderSize = 48

func (h *Header) IsFpkd() bool {
	return bytes.Compare(h.Magic[:], MagicFpkd[:]) == 0
}

func (h *Header) IsValid() bool {
	return bytes.Compare(h.Magic[:], MagicFpkd[:]) == 0 || bytes.Compare(h.Magic[:], MagicFpk[:]) == 0
}

func (h *Header) Read(reader io.Reader) error {
	if err := binary.Read(reader, binary.LittleEndian, h); err != nil {
		return fmt.Errorf("%w", err)
	}

	if !h.IsValid() {
		return fmt.Errorf("unknown fpk(d) magic: [% x]", h.Magic)
	}

	return nil
}

func (h *Header) Write(writer io.Writer) error {
	h.MagicNumber2 = 2
	return binary.Write(writer, binary.LittleEndian, h)
}

func (h *Header) SetType(isFpkd bool) {
	if isFpkd {
		h.Magic = MagicFpkd
	} else {
		h.Magic = MagicFpk
	}
}
