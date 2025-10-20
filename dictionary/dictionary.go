package dictionary

import (
	"bytes"
	"io"

	"github.com/unknown321/hashing"
)

type DictStrCode64 struct {
	data map[uint32]string
}

func (d *DictStrCode64) Read(reader io.Reader) error {
	var b []byte
	var err error
	if b, err = io.ReadAll(reader); err != nil {
		return err
	}

	lines := bytes.Split(b, []byte("\r\n"))
	if len(lines) < 2 {
		lines = bytes.Split(b, []byte("\n"))
	}

	d.data = make(map[uint32]string)

	for _, line := range lines {
		h := uint32(hashing.StrCode64(line) & 0xffffffff)
		if _, ok := d.data[h]; !ok {
			d.data[h] = string(line)
		}
	}

	return nil
}

func (d *DictStrCode64) Get(hash uint32) string {
	if v, ok := d.data[hash]; ok {
		return v
	}

	return ""
}
