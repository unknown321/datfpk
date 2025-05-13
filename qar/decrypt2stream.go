package qar

import (
	"encoding/binary"
	"fmt"
	"io"
)

type Decrypt2Stream struct {
	Key      uint32
	BlockKey uint32
}

func (d *Decrypt2Stream) Init(key uint32) {
	d.Key = key * 278
	d.BlockKey = key | ((key ^ 25974) << 16)
	//slog.Debug("d2init", "key", key, "blockkey", d.BlockKey)
}

func (d *Decrypt2Stream) Read(reader io.Reader, count int) ([]byte, error) {
	buf := make([]byte, count)
	n, err := reader.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("cannot read decrypt2stream: %w", err)
	}

	res := make([]byte, count)
	if n < 1 {
		return nil, fmt.Errorf("nothing to read")
	}

	//slog.Debug("preDecrypt", "v", fmt.Sprintf("% x, %s", buf, buf), "n", n)

	res, err = d.Decrypt2(buf, n)
	if err != nil {
		return nil, fmt.Errorf("decrypt2 fail: %w", err)
	}

	//slog.Debug("postDecrypt", "v", fmt.Sprintf("% x", res), "n", n)

	return res, nil
}

func (d *Decrypt2Stream) Decrypt2(input []byte, size int) ([]byte, error) {
	output := make([]byte, len(input))
	copy(output, input)

	pDest := output
	pSrc := input

	for size >= 64 {
		for j := 16; j > 0; j-- {
			x := d.BlockKey ^ binary.LittleEndian.Uint32(pSrc)
			binary.LittleEndian.PutUint32(pDest, x)
			d.BlockKey = d.Key + 48828125*d.BlockKey
			pSrc = pSrc[4:]
			pDest = pDest[4:]
		}
		size -= 64
	}

	for size >= 16 {
		x := d.BlockKey ^ binary.LittleEndian.Uint32(pSrc)
		v7 := d.Key + 48828125*d.BlockKey
		d2 := v7 ^ binary.LittleEndian.Uint32(pSrc[4:])
		v8 := d.Key + 48828125*v7
		d3 := v8 ^ binary.LittleEndian.Uint32(pSrc[8:])
		v9 := d.Key + 48828125*v8
		d4 := v9 ^ binary.LittleEndian.Uint32(pSrc[12:])

		binary.LittleEndian.PutUint32(pDest, x)
		binary.LittleEndian.PutUint32(pDest[4:], d2)
		binary.LittleEndian.PutUint32(pDest[8:], d3)
		binary.LittleEndian.PutUint32(pDest[12:], d4)

		d.BlockKey = d.Key + 48828125*v9
		size -= 16
		pSrc = pSrc[16:]
		pDest = pDest[16:]
	}

	for size >= 4 {
		x := d.BlockKey ^ binary.LittleEndian.Uint32(pSrc)
		binary.LittleEndian.PutUint32(pDest, x)
		d.BlockKey = d.Key + 48828125*d.BlockKey
		size -= 4
		pSrc = pSrc[4:]
		pDest = pDest[4:]
	}

	// The final 0-3 bytes aren't encrypted, leave them as is
	return output, nil
}
