package hashing

import (
	"bytes"
	"fmt"
	"io"
)

type Dictionary struct {
	Hashes     map[uint64]string
	Extensions map[uint64]string
}

func (d *Dictionary) Read(reader io.Reader) error {
	var b []byte
	var err error
	if b, err = io.ReadAll(reader); err != nil {
		return err
	}

	lines := bytes.Split(b, []byte("\r\n"))
	if len(lines) < 2 {
		lines = bytes.Split(b, []byte("\n"))
	}

	d.Hashes = make(map[uint64]string)
	d.Extensions = make(map[uint64]string)

	for _, line := range lines {
		h := HashFileName(string(line), true) & 0x3FFFFFFFFFFFF
		if _, ok := d.Hashes[h]; !ok {
			d.Hashes[h] = string(line)
		}
	}

	return nil
}

func (d *Dictionary) GetByHash(hash uint64) (string, bool) {
	ext := ExtHashFromHash(hash)
	path := PathHashFromHash(hash)
	//slog.Debug("getting by hash", "hash", fmt.Sprintf("%x", hash), "path", fmt.Sprintf("%x", path), "ext", fmt.Sprintf("%x", ext))
	resolved := true
	pp, ok := d.Hashes[path]
	if !ok {
		pp = fmt.Sprintf("%x", path)
		resolved = false
	}

	ee, ok := ExtensionsByHash[ext]
	if !ok {
		ee = fmt.Sprintf("%x.unknown", ext)
		resolved = false
	}
	return fmt.Sprintf("%s.%s", pp, ee), resolved
}

func (d *Dictionary) CalcHash(s []byte) uint64 {
	return StrCode32(s)
}
