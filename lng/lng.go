package lng

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"slices"

	"github.com/unknown321/datfpk/dictionary"
	"github.com/unknown321/hashing"
)

const LngID = "Lng"

type Lng struct {
	Header  Header
	Entries []Entry
	Keys    []Key
}

type entryJson struct {
	Key    entryKey `json:"key,omitempty"`
	LangId string   `json:"lang_id,omitempty"`
	Color  int16    `json:"color"`
	Value  string   `json:"value"`
}

type lngJson struct {
	Type       string      `json:"type"`
	Version    Version     `json:"version"`
	Endianness string      `json:"endianness"`
	Entries    []entryJson `json:"entries"`
}

func (l *Lng) Read(seeker io.ReadSeeker, dictionary dictionary.DictStrCode64) error {
	var err error
	if err = l.Header.Read(seeker); err != nil {
		return fmt.Errorf("header read error: %w", err)
	}

	if _, err = seeker.Seek(int64(l.Header.ValuesOffset), io.SeekStart); err != nil {
		return fmt.Errorf("seek to values: %w", err)
	}

	var endianness binary.ByteOrder
	if l.Header.Endianness == EndiannessLE {
		endianness = binary.LittleEndian
	} else {
		endianness = binary.BigEndian
	}

	var i uint32 = 0
	for i = 0; i < l.Header.EntryCount; i++ {
		e := Entry{}
		if err = e.Read(seeker, endianness); err != nil {
			return fmt.Errorf("entry read error: %w", err)
		}

		e.Offset -= int64(l.Header.ValuesOffset)

		l.Entries = append(l.Entries, e)
	}

	if _, err = seeker.Seek(int64(l.Header.KeysOffset), io.SeekStart); err != nil {
		return fmt.Errorf("seek to keys: %w", err)
	}

	for i = 0; i < l.Header.EntryCount; i++ {
		k := Key{}
		if err = k.Read(seeker, endianness); err != nil {
			return fmt.Errorf("key read error: %w", err)
		}

		l.Keys = append(l.Keys, k)
	}

	for i = 0; i < l.Header.EntryCount; i++ {
		var k uint32
		for k = 0; k < l.Header.EntryCount; k++ {
			if l.Entries[i].Offset == int64(l.Keys[k].Offset) {
				l.Entries[i].LangId = dictionary.Get(l.Keys[k].Key)
				l.Entries[i].Key = l.Keys[k].Key
				break
			}
		}
	}

	return nil
}

func (l *Lng) MarshalJSON() ([]byte, error) {
	entries := make([]entryJson, len(l.Entries))

	for i, entry := range l.Entries {
		entries[i].Color = entry.Color
		entries[i].Value = entry.Value
		entries[i].LangId = entry.LangId
		if entry.LangId == "" {
			entries[i].Key = entryKey(entry.Key)
		}
	}

	end := make([]byte, 4)
	binary.LittleEndian.PutUint32(end, uint32(l.Header.Endianness))
	ll := lngJson{
		Type:       LngID,
		Version:    l.Header.Version,
		Endianness: string(end[0:2]),
		Entries:    entries,
	}

	var resultBytes bytes.Buffer
	enc := json.NewEncoder(&resultBytes)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(ll); err != nil {
		return nil, fmt.Errorf("json marshal error: %w", err)
	}

	return resultBytes.Bytes(), nil
}

func (l *Lng) UnmarshalJSON(b []byte) error {
	ll := lngJson{}
	if err := json.Unmarshal(b, &ll); err != nil {
		return err
	}

	l.Header.Version = ll.Version
	if ll.Endianness == "LE" {
		l.Header.Endianness = EndiannessLE
	}

	if ll.Endianness == "BE" {
		l.Header.Endianness = EndiannessBE
	}

	if l.Header.Endianness == 0 {
		return fmt.Errorf("invalid endianness (expected LE or BE)")
	}

	for _, entry := range ll.Entries {
		e := Entry{
			LangId: entry.LangId,
			Offset: 0,
			Color:  entry.Color,
			Value:  entry.Value,
			Key:    uint32(entry.Key),
		}

		l.Entries = append(l.Entries, e)
	}

	return nil
}

func (l *Lng) Write(seeker io.WriteSeeker) error {
	var err error
	var endianness binary.ByteOrder

	if l.Header.Endianness == EndiannessBE || l.Header.Endianness == 0 {
		endianness = binary.BigEndian
	} else {
		endianness = binary.LittleEndian
	}

	l.Header.ValuesOffset = HeaderSize
	if _, err = seeker.Seek(HeaderSize, io.SeekStart); err != nil {
		return fmt.Errorf("seek to data: %w", err)
	}

	l.Keys = make([]Key, len(l.Entries))

	for i, entry := range l.Entries {
		offset, err := seeker.Seek(0, io.SeekCurrent)
		if err != nil {
			return fmt.Errorf("get entry offset: %w", err)
		}

		if err = entry.Write(seeker, endianness); err != nil {
			return fmt.Errorf("entry write error: %w", err)
		}

		l.Keys[i].Offset = uint32(offset) - HeaderSize
		if entry.Key == 0 {
			l.Keys[i].Key = uint32(hashing.StrCode64([]byte(entry.LangId)) & 0xffffffff)
		} else {
			l.Keys[i].Key = entry.Key
		}
	}

	offset, err := seeker.Seek(0, io.SeekCurrent)
	if err != nil {
		return fmt.Errorf("get entry offset: %w", err)
	}

	paddingSize := 4 - offset%4
	padding := make([]byte, paddingSize)
	if _, err = seeker.Write(padding); err != nil {
		return fmt.Errorf("padding: %w", err)
	}

	l.Header.KeysOffset = int32(offset + paddingSize)
	slices.SortFunc(l.Keys, func(a, b Key) int {
		if a.Key < b.Key {
			//if a.Offset < b.Offset {
			//	return 1
			//}
			return -1
		}

		return 1
	})

	for _, key := range l.Keys {
		if err = key.Write(seeker, endianness); err != nil {
			return fmt.Errorf("key write error: %w", err)
		}
	}

	if _, err = seeker.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("seek to start: %w", err)
	}

	l.Header.EntryCount = uint32(len(l.Entries))
	l.Header.Magic = Magic

	if err = l.Header.Write(seeker, endianness); err != nil {
		return fmt.Errorf("header write error: %w", err)
	}

	return nil
}
