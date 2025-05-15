package fpk

import (
	"bytes"
	"crypto/md5"
	"datfpk/hashing"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"unsafe"
)

type Entry struct {
	DataOffset uint32
	DataSize   uint32
	FilePath   String
	PathMD5    [16]byte // md5.Sum(e.FilePath.Data)
	Data       []byte
	Encrypted  bool
}

type ejs struct {
	FilePath  string `json:"filePath"`
	Encrypted bool   `json:"encrypted,omitempty"`
}

func (e *Entry) MarshalJSON() ([]byte, error) {
	j := ejs{
		FilePath:  e.FilePath.Data,
		Encrypted: e.Encrypted,
	}

	return json.Marshal(j)
}

func (e *Entry) UnmarshalJSON(i []byte) error {
	j := ejs{}
	if err := json.Unmarshal(i, &j); err != nil {
		return err
	}
	e.FilePath.Data = j.FilePath
	e.Encrypted = j.Encrypted
	e.PathMD5 = md5.Sum([]byte(e.FilePath.Data))

	return nil
}

const EntrySize = 4*4 + 4*4 + 16 // dataInfo, filePath, md5

func (e *Entry) Read(reader io.ReadSeeker) error {
	//o, _ := reader.Seek(0, io.SeekCurrent)
	//slog.Info("entry", "offset", o)
	skip := uint32(0)
	binary.Read(reader, binary.LittleEndian, &e.DataOffset)
	binary.Read(reader, binary.LittleEndian, skip)
	binary.Read(reader, binary.LittleEndian, &e.DataSize)
	binary.Read(reader, binary.LittleEndian, skip)
	if err := e.FilePath.Read(reader); err != nil {
		return fmt.Errorf("%w", err)
	}
	reader.Read(e.PathMD5[:])

	//slog.Info("entry",
	//	"md5", fmt.Sprintf("%x", e.PathMD5),
	//	"path", fmt.Sprintf("%x", md5.Sum([]byte(e.FilePath.Data))),
	//)

	if err := e.ReadData(reader); err != nil {
		return fmt.Errorf("read: %w", err)
	}

	return nil
}

func (e *Entry) ReadData(reader io.ReadSeeker) error {
	if e.DataSize < 1 {
		return nil
	}

	var curPos int64
	var err error
	if curPos, err = reader.Seek(0, io.SeekCurrent); err != nil {
		return fmt.Errorf("curpos: %w", err)
	}
	//slog.Info("data", "offset", curPos)

	_, err = reader.Seek(int64(e.DataOffset), io.SeekStart)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	b := make([]byte, e.DataSize)
	_, err = reader.Read(b)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if b[0] == 0x1B || b[0] == 0x1C {
		//slog.Info("enc", "q", e.DataSize, "o", e.DataOffset, "s", e.FilePath.Data)
		if dec, err := Decrypt(b, e.FilePath.Data); err != nil {
			// decrypt failed
			e.Data = b
			err = nil
		} else {
			e.Data = dec
			e.Encrypted = true
		}
	} else {
		e.Data = b
	}

	_, _ = reader.Seek(curPos, io.SeekStart)

	return nil
}

const blockSize = int(unsafe.Sizeof(uint64(0)))

func Decrypt(data []byte, name string) ([]byte, error) {
	fName := filepath.Base(strings.ToLower(name))
	h := hashing.HashFileNameLegacy([]byte(fName), false)
	key := make([]byte, blockSize)
	binary.LittleEndian.PutUint64(key, ^h)

	//slog.Info("decrypt", "name", fName, "hash", h, "key", fmt.Sprintf("%x", key))

	res := make([]byte, len(data)-1)
	for i := 0; i < len(data)-1; i++ {
		//if i < 16 {
		//	slog.Info("q", "i", i,
		//		"data", fmt.Sprintf("% x (%s)", data[i+1], string(data[i+1])),
		//		"key", fmt.Sprintf("% x (%s)", key[i%blockSize], string(key[i%blockSize])),
		//		"res", fmt.Sprintf("% x (%s)", key[i%blockSize]^data[i+1], string(key[i%blockSize]^data[i+1])))
		//}
		key[i%blockSize] ^= data[i+1]
		res[i] = key[i%blockSize]
	}

	if res[len(res)-1] != 0 {
		return res, fmt.Errorf("last byte is not null: % x", res[len(res)-1])
	}

	return res[:len(res)-1], nil
}

func Encrypt(data []byte, name string) []byte {
	fName := filepath.Base(strings.ToLower(name))
	h := hashing.HashFileNameLegacy([]byte(fName), false)
	key := make([]byte, blockSize)
	binary.LittleEndian.PutUint64(key, ^h)
	data = append(data, 0x0)

	//slog.Info("encrypt", "name", fName, "hash", h, "key", fmt.Sprintf("%x", key))

	res := make([]byte, len(data))
	for i := 0; i < len(data); i++ {
		//if i < 16 && i > 7 {
		//	slog.Info("q", "i", i,
		//		"data", fmt.Sprintf("% x (%s)", data[i+1], string(data[i+1])),
		//		"key", fmt.Sprintf("% x (%s)", key[i%blockSize], string(key[i%blockSize])),
		//		"res", fmt.Sprintf("% x (%s)", key[i%blockSize]^data[i+1], string(key[i%blockSize]^data[i+1])))
		//}
		r := key[i%blockSize] ^ data[i]
		res[i] = r
		key[i%blockSize] = data[i]
	}

	res = append([]byte{0x1B}, res...)
	return res
}

func (e *Entry) WriteData(writer io.WriteSeeker, name string) error {
	if e.Encrypted {
		e.Data = Encrypt(e.Data, name)
	}

	var err error
	var o int64
	if o, err = writer.Seek(0, io.SeekCurrent); err != nil {
		return err
	}

	e.DataOffset = uint32(o)
	e.DataSize = uint32(len(e.Data))

	if _, err = writer.Write(e.Data); err != nil {
		return err
	}

	return nil
}

func (e *Entry) WriteHeader(writer io.WriteSeeker) error {
	skip := uint32(0)
	binary.Write(writer, binary.LittleEndian, e.DataOffset)
	binary.Write(writer, binary.LittleEndian, skip)
	binary.Write(writer, binary.LittleEndian, e.DataSize)
	binary.Write(writer, binary.LittleEndian, skip)
	e.FilePath.Header.Write(writer)
	if md5empty(e.PathMD5) {
		e.PathMD5 = md5.Sum([]byte(e.FilePath.Data))
	}
	writer.Write(e.PathMD5[:])

	return nil
}

func md5empty(data [16]byte) bool {
	empty := make([]byte, 16)
	return bytes.Compare(empty, data[:]) == 0
}
