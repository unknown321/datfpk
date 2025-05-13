package util

import (
	"errors"
	"io"
)

// TODO WriteSeeker -> Writer

func AlignWrite(writer io.WriteSeeker, alignment int64) (int64, error) {
	pos, err := writer.Seek(0, io.SeekCurrent)
	if pos%alignment != 0 {
		align := make([]byte, alignment-pos%alignment)
		if _, err = writer.Write(align); err != nil {
			return 0, err
		}
	}

	if pos, err = writer.Seek(0, io.SeekCurrent); err != nil {
		return 0, err
	}

	return pos, nil
}

func AlignRead(reader io.ReadSeeker, alignment int64) (int64, error) {
	pos, err := reader.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, err
	}

	if pos%alignment != 0 {
		if pos, err = reader.Seek(alignment-pos%alignment, io.SeekCurrent); err != nil {
			return 0, err
		}
	}

	return pos, nil
}

// ByteArrayReaderWriter used for testing
type ByteArrayReaderWriter struct {
	data []byte
	pos  int64
}

func NewByteArrayReaderWriter(data []byte) *ByteArrayReaderWriter {
	return &ByteArrayReaderWriter{
		data: data,
		pos:  0,
	}
}

func (b *ByteArrayReaderWriter) Read(p []byte) (n int, err error) {
	if b.pos >= int64(len(b.data)) {
		return 0, io.EOF
	}
	n = copy(p, b.data[b.pos:])
	b.pos += int64(n)
	return n, nil
}

func (b *ByteArrayReaderWriter) Write(p []byte) (n int, err error) {
	if b.pos+int64(len(p)) > int64(cap(b.data)) {
		newData := make([]byte, b.pos+int64(len(p)))
		copy(newData, b.data)
		b.data = newData
	}
	n = copy(b.data[b.pos:], p)
	b.pos += int64(n)
	if b.pos > int64(len(b.data)) {
		b.data = b.data[:b.pos]
	}
	return n, nil
}

func (b *ByteArrayReaderWriter) Bytes() []byte {
	return b.data
}

func (b *ByteArrayReaderWriter) Seek(offset int64, whence int) (int64, error) {
	var newPos int64
	switch whence {
	case io.SeekStart:
		newPos = offset
	case io.SeekCurrent:
		newPos = b.pos + offset
	case io.SeekEnd:
		newPos = int64(len(b.data)) + offset
	default:
		return 0, errors.New("invalid whence")
	}
	if newPos < 0 {
		return 0, errors.New("invalid position")
	}
	b.pos = newPos
	return b.pos, nil
}

// CompactStringSlice is copy of slices.Compact, available in go1.21
func CompactStringSlice(s []string) []string {
	if len(s) < 2 {
		return s
	}
	for k := 1; k < len(s); k++ {
		if s[k] == s[k-1] {
			s2 := s[k:]
			for k2 := 1; k2 < len(s2); k2++ {
				if s2[k2] != s2[k2-1] {
					s[k] = s2[k2]
					k++
				}
			}

			clear(s[k:]) // zero/nil out the obsolete elements, for GC
			return s[:k]
		}
	}
	return s
}
