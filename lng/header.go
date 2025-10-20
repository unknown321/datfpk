package lng

import (
	"encoding/binary"
	"fmt"
	"io"
)

const HeaderSize = 24
const Magic = 0x474E414C // LANG

type Endianness int32

const (
	EndiannessLE Endianness = 0x454C
	EndiannessBE            = 0x4542
)

type Version uint32

const (
	VersionGZ  Version = 2
	VersionTPP Version = 3
)

type Header struct {
	Magic        uint32     `json:"-"`
	Version      Version    `json:"version"`
	Endianness   Endianness `json:"endianness"`
	EntryCount   uint32     `json:"entry_count"`
	ValuesOffset int32      `json:"-"`
	KeysOffset   int32      `json:"-"`
}

func (h *Header) Read(seeker io.ReadSeeker) error {
	var err error
	err = binary.Read(seeker, binary.LittleEndian, &h.Magic)
	if err != nil {
		return fmt.Errorf("magic: %w", err)
	}

	if _, err = seeker.Seek(8, io.SeekStart); err != nil {
		return fmt.Errorf("seek to endianness: %w", err)
	}

	if err = binary.Read(seeker, binary.LittleEndian, &h.Endianness); err != nil {
		return fmt.Errorf("endianness: %w", err)
	}

	if h.Endianness != EndiannessLE && h.Endianness != EndiannessBE {
		return fmt.Errorf("unknown endianness: %x", h.Endianness)
	}

	var endianness binary.ByteOrder
	if h.Endianness == EndiannessBE {
		endianness = binary.BigEndian
	} else {
		endianness = binary.LittleEndian
	}

	if _, err = seeker.Seek(4, io.SeekStart); err != nil {
		return fmt.Errorf("seek to version: %w", err)
	}

	if err = binary.Read(seeker, endianness, &h.Version); err != nil {
		return fmt.Errorf("version: %w", err)
	}

	if h.Version != VersionGZ && h.Version != VersionTPP {
		return fmt.Errorf("header: invalid version (%d)", h.Version)
	}

	if _, err = seeker.Seek(12, io.SeekStart); err != nil {
		return fmt.Errorf("seek to entry count: %w", err)
	}

	if err = binary.Read(seeker, endianness, &h.EntryCount); err != nil {
		return fmt.Errorf("entry count: %w", err)
	}

	if err = binary.Read(seeker, endianness, &h.ValuesOffset); err != nil {
		return fmt.Errorf("values offset: %w", err)
	}

	if err = binary.Read(seeker, endianness, &h.KeysOffset); err != nil {
		return fmt.Errorf("keys offset: %w", err)
	}

	return nil
}

func (h *Header) Write(w io.Writer, order binary.ByteOrder) error {
	var err error
	if err = binary.Write(w, binary.LittleEndian, h.Magic); err != nil {
		return fmt.Errorf("magic: %w", err)
	}

	if err = binary.Write(w, order, h.Version); err != nil {
		return fmt.Errorf("version: %w", err)
	}

	if err = binary.Write(w, binary.LittleEndian, h.Endianness); err != nil {
		return fmt.Errorf("endianness: %w", err)
	}

	if err = binary.Write(w, order, h.EntryCount); err != nil {
		return fmt.Errorf("entry count: %w", err)
	}

	if err = binary.Write(w, order, h.ValuesOffset); err != nil {
		return fmt.Errorf("values offset: %w", err)
	}

	if err = binary.Write(w, order, h.KeysOffset); err != nil {
		return fmt.Errorf("keys offset: %w", err)
	}

	return nil
}
