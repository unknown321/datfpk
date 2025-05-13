package qar

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

type Md5Sum [16]byte

func (d *Md5Sum) UnmarshalJSON(i []byte) error {
	s := strings.ReplaceAll(string(i), "\"", "")
	if s == "0" || s == "" {
		for n := range d {
			d[n] = 0
		}

		return nil
	}

	v, err := hex.DecodeString(s)
	if err != nil {
		return err
	}

	if len(v) != 16 {
		return fmt.Errorf("md5sum is not 16 bytes")
	}

	for n, r := range v {
		d[n] = r
	}

	return nil
}

func (d *Md5Sum) Empty() bool {
	empty := make([]byte, 16)
	return bytes.Compare(d[:], empty) == 0
}

func (d *Md5Sum) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(d[:]))
}

func Md5Decode(in Md5Sum) (Md5Sum, error) {
	reader := bytes.NewReader(in[:])
	md51, err := ReadUint32(reader)
	if err != nil {
		return Md5Sum{}, err
	}
	md51 ^= xorMask4

	md52, err := ReadUint32(reader)
	if err != nil {
		return Md5Sum{}, err
	}
	md52 ^= xorMask1

	md53, err := ReadUint32(reader)
	if err != nil {
		return Md5Sum{}, err
	}
	md53 ^= xorMask1

	md54, err := ReadUint32(reader)
	if err != nil {
		return Md5Sum{}, err
	}
	md54 ^= xorMask2

	out := Md5Sum{}
	binary.LittleEndian.PutUint32(out[0:], md51)
	binary.LittleEndian.PutUint32(out[4:], md52)
	binary.LittleEndian.PutUint32(out[8:], md53)
	binary.LittleEndian.PutUint32(out[12:], md54)

	return out, nil
}
