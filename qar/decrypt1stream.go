package qar

import (
	"encoding/binary"
	"fmt"
	"io"
	"unsafe"
)

var decryptionTable = []uint32{
	0xBB8ADEDB,
	0x65229958,
	0x08453206,
	0x88121302,
	0x4C344955,
	0x2C02F10C,
	0x4887F823,
	0xF3818583,
}

const blockSize = int(unsafe.Sizeof(uint64(0)))
const blockSize2 = int(unsafe.Sizeof(uint32(0)))

type Decrypt1Stream struct {
	Size     int
	Position int
	Version  int
	HashLow  uint32
	Seed     uint64
	SeedLow  uint32
	SeedHigh uint32
}

func (d *Decrypt1Stream) Init(md5sum Md5Sum, pathHash uint64, version uint32, size int) {
	hashLow := uint32(pathHash & 0xFFFFFFFF)
	d.HashLow = hashLow
	d.Seed = binary.LittleEndian.Uint64(md5sum[int(hashLow%2)*8:])
	d.SeedLow = uint32(d.Seed & 0xFFFFFFFF)
	d.SeedHigh = uint32(d.Seed >> 32)
	d.Version = int(version)
	d.Size = size

	//slog.Debug("d1 init", "hashlow", d.HashLow, "seed", d.Seed, "seedLow", d.SeedLow, "seedHigh", d.SeedHigh, "version", d.Version, "size", d.Size)
}

func (d *Decrypt1Stream) Read(reader io.Reader, count int) ([]byte, error) {
	if count > (d.Size - d.Position) {
		count = d.Size - d.Position
	}

	pad := 8 - count%8

	buf := make([]byte, count+pad)
	n, err := reader.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("cannot read decrypt1stream: %w", err)
	}

	//slog.Debug("d1", "d1", fmt.Sprintf("% d", buf[0:8]))
	if n == 0 {
		return nil, nil
	}

	d.Decrypt1(buf)

	buf = buf[0:count]
	//slog.Debug("d1stream", "result", fmt.Sprintf("% x", buf))

	d.Position += n

	return buf, nil
}

func (d *Decrypt1Stream) Decrypt1(data []byte) {
	blocks := len(data) / blockSize
	//slog.Debug("blocks", "c", blocks, "ld", len(data))

	if d.Version == 2 {
		for i := 0; i < blocks; i++ {
			offset1 := i * blockSize
			offset2 := i*blockSize + blockSize2
			offset1Absolute := offset1 + d.Position

			index := 2 * int((uint64(d.HashLow)+d.Seed+uint64(offset1Absolute/11))%4)
			u1 := binary.LittleEndian.Uint32(data[offset1:]) ^ decryptionTable[index] ^ d.SeedLow
			u2 := binary.LittleEndian.Uint32(data[offset2:]) ^ decryptionTable[index+1] ^ d.SeedHigh
			binary.LittleEndian.PutUint32(data[offset1:], u1)
			binary.LittleEndian.PutUint32(data[offset2:], u2)
		}

		remaining := len(data) % blockSize
		for i := 0; i < remaining; i++ {
			offset := blocks*blockSize + i
			offsetBlock := offset - (offset % blockSize)
			offsetBlockAbsolute := offsetBlock + d.Position

			index := 2 * int((uint64(d.HashLow)+d.Seed+uint64(offsetBlockAbsolute/11))%4)
			decryptionIndex := offset % blockSize

			xorMask := decryptionTable[index+1]
			seedMask := d.SeedHigh
			if decryptionIndex < 4 {
				xorMask = decryptionTable[index]
				seedMask = d.SeedLow
			}

			xorMaskByte := byte((xorMask >> (8 * (decryptionIndex % 4))) & 0xff)
			seedByte := (byte)((seedMask >> (8 * (decryptionIndex % 4))) & 0xff)
			data[offset] = data[offset] ^ (xorMaskByte ^ seedByte)
		}

		return
	}

	//slog.Debug("decrypt1", "hashlow", d.HashLow, "data", fmt.Sprintf("% x", data))
	for i := 0; i < blocks; i++ {
		offset1 := i * blockSize
		offset2 := i*blockSize + blockSize2
		offset1Absolute := offset1 + d.Position

		index := 2 * int((uint64(d.HashLow)+uint64(offset1Absolute/11))%4)
		//slog.Debug("i", "index", index, "off1", offset1, "off2", offset2)
		u1 := binary.LittleEndian.Uint32(data[offset1:]) ^ decryptionTable[index]
		u2 := binary.LittleEndian.Uint32(data[offset2:]) ^ decryptionTable[index+1]
		binary.LittleEndian.PutUint32(data[offset1:], u1)
		binary.LittleEndian.PutUint32(data[offset2:], u2)
	}

	// in practice this code never runs
	remaining := len(data) % blockSize
	for i := 0; i < remaining; i++ {
		offset := blocks*blockSize + i
		offsetAbsolute := offset + d.Position
		index := int(2 * ((d.HashLow + uint32(offsetAbsolute-(offsetAbsolute%blockSize))) / 11) % 4)
		decryptionIndex := offset % blockSize
		xorMask := decryptionTable[index+1]
		if decryptionIndex < 4 {
			xorMask = decryptionTable[index]
		}
		xorMaskByte := byte((xorMask >> (8 * decryptionIndex)) & 0xff)
		b1 := data[offset] ^ xorMaskByte
		data[offset] = b1
	}
}
