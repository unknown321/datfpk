package qar

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/unknown321/datfpk/util"

	"github.com/unknown321/hashing"
)

type EntryFromDisk struct {
	Path       string
	BaseDir    string
	Compressed bool
	MetaFlag   bool
	Encrypted  bool
	Key        uint32
}

var dataDir = "testdata"

func TestQar_ExtractTo(t *testing.T) {
	data := map[string][]byte{
		//"encrypted.dat":  nil, // MGSV_QAR_TOOL and GzsTool don't encrypt files (or do they?), so there is no way to create test qar with encrypted data using other tools
		"plain.dat":      []byte("return 'data1234567890'\n"),
		"compressed.dat": []byte("data1234567890\ndata1234567\n"), // there is a limit on minimum data size, >16 bytes compressed?
		"ps3.dat":        []byte("return 'data1234567890'\n"),
	}

	for k, d := range data {
		t.Run(k, func(t *testing.T) {
			q := Qar{}

			fp := filepath.Join(".", dataDir, k)

			if err := q.ReadFrom(fp); err != nil {
				t.Errorf("%s", err.Error())
			}
			defer q.Close()

			name := "/test.lua"
			hash := hashing.HashFileNameWithExtension(name)
			b := new(bytes.Buffer)
			if _, err := q.ExtractTo(name, hash, b); err != nil {
				t.Errorf("%s", err.Error())
			}

			if bytes.Compare(b.Bytes(), d) != 0 {
				t.Errorf("not equal: %s != %s", b.Bytes(), d)
			}
		})
	}
}

// TestQar_WriteThenRead checks if we can read whatever we wrote and that data matches
func TestQar_WriteThenRead(t *testing.T) {
	type fields struct {
		Flags           uint32
		UnknownCount    uint32
		Version         uint32
		IsConsole       bool
		EntriesFromDisk []EntryFromDisk
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "plain",
			fields: fields{
				Flags:        3150304,
				UnknownCount: 0,
				Version:      1,
				EntriesFromDisk: []EntryFromDisk{{
					BaseDir:    "plain_dat",
					Path:       "test.lua",
					Compressed: false,
					MetaFlag:   false,
				}},
			},
			wantErr: false,
		},
		{
			name: "compressed",
			fields: fields{
				Flags:        3150304,
				UnknownCount: 0,
				Version:      1,
				EntriesFromDisk: []EntryFromDisk{{
					BaseDir:    "compressed_dat",
					Path:       "test.lua",
					Compressed: true,
					MetaFlag:   false,
				}},
			},
			wantErr: false,
		},
		{
			name: "ps3",
			fields: fields{
				Flags:        0,
				UnknownCount: 0,
				Version:      0x24,
				IsConsole:    true,
				EntriesFromDisk: []EntryFromDisk{{
					BaseDir:    "ps3_dat",
					Path:       "test.lua",
					Compressed: false,
					MetaFlag:   false,
				}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &Qar{
				Magic:        PCMagic,
				Flags:        tt.fields.Flags,
				UnknownCount: tt.fields.UnknownCount,
				Version:      tt.fields.Version,
				isConsole:    tt.fields.IsConsole,
				Entries:      []Entry{},
			}

			if tt.fields.IsConsole {
				q.Magic = ConsoleMagic
				q.isConsole = true
			}

			// add qar entries
			var err error
			for _, v := range tt.fields.EntriesFromDisk {
				e := Entry{
					Header: EntryHeader{
						FilePath:   v.Path,
						Compressed: v.Compressed,
						MetaFlag:   v.MetaFlag,
					},
					DataHeader: DataHeader{
						EncryptionMagic: 0,
						Key:             0,
					},
					Data: nil,
				}

				p := filepath.Join(dataDir, v.BaseDir, v.Path)

				if e.Data, err = os.ReadFile(p); err != nil {
					t.Errorf("%s", err.Error())
				}
				e.Header.UncompressedSize = uint32(len(e.Data))
				q.Entries = append(q.Entries, e)
			}
			q.FileCount = uint32(len(q.Entries))

			out := &util.ByteArrayReaderWriter{}
			if err = q.Write(out, "", false); (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
			}

			ourData := &Qar{}
			br := util.NewByteArrayReaderWriter(out.Bytes())
			if err = ourData.Read(br); err != nil {
				t.Fatalf("cannot read generated qar: %s", err.Error())
			}

			if err = ourData.Entries[0].ReadData(br); err != nil {
				t.Fatalf("%s", err.Error())
			}

			v := tt.fields.EntriesFromDisk[0]
			p := filepath.Join(dataDir, v.BaseDir, v.Path)
			referenceData, err := os.ReadFile(p)
			if err != nil {
				t.Fatalf("%s", err.Error())
			}

			// console Read does not read data from qar
			if !q.isConsole {
				if bytes.Compare(referenceData, ourData.Entries[0].Data) != 0 {
					t.Errorf("not equal, want %s from %s, have % x", referenceData, p, ourData.Entries[0].Data)
				}
			}
		})
	}
}

func TestQar_ExtractByHash(t *testing.T) {
	var err error
	q := Qar{}
	if err = q.ReadFrom("testdata/plain.dat"); err != nil {
		t.Fatalf("%s", err.Error())
	}

	l := ""
	hash := uint64(0)
	dict := hashing.Dictionary{}
	for n, v := range q.Entries {
		entryName, ok := dict.GetByHash(v.Header.PathHash)
		if ok {
			q.Entries[n].Header.FilePath = entryName
		} else {
			q.Entries[n].Header.FilePath = entryName
		}

		l = fmt.Sprintf("%x", v.Header.PathHash)
		hash = v.Header.PathHash
	}

	b := []byte{}
	w := bytes.NewBuffer(b)

	if _, err = q.ExtractTo(l, hash, w); err != nil {
		t.Fatalf("%s", err.Error())
	}
}

func TestDecryptSectionList(t *testing.T) {
	section := uint64(0x123456789)
	sections := make([]byte, blockSize*1)
	binary.LittleEndian.PutUint64(sections, section)
	enc, err := DecryptSectionList(1, sections, 1, true)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	binary.LittleEndian.PutUint64(sections, enc[0])
	dec, err := DecryptSectionList(1, sections, 1, false)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	if dec[0] != section {
		t.Fatalf("not equal")
	}
}

// TestQar_Write checks if a qar generated in memory from testdata matches a pre-generated qar file on disk
func TestQar_Write(t *testing.T) {
	type fields struct {
		Flags           uint32
		FileCount       uint32
		UnknownCount    uint32
		Version         uint32
		IsConsole       bool
		EntriesFromDisk []EntryFromDisk
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "plain",
			fields: fields{
				Flags:        3150304,
				FileCount:    1,
				UnknownCount: 0,
				Version:      1,
				EntriesFromDisk: []EntryFromDisk{{
					BaseDir:    "plain_dat",
					Path:       "test.lua",
					Compressed: false,
					MetaFlag:   false,
				}},
			},
			wantErr: false,
		},
		{
			name: "ps3",
			fields: fields{
				Flags:        0,
				FileCount:    1,
				UnknownCount: 0,
				Version:      0x24,
				IsConsole:    true,
				EntriesFromDisk: []EntryFromDisk{{
					BaseDir:    "ps3_dat",
					Path:       "test.lua",
					Compressed: false,
					MetaFlag:   false,
					Encrypted:  true,
					Key:        0x8eb31c0f,
				}},
			},
			wantErr: false,
		},
		/*
				// this test won't work due to differences in zlib implementation
			    // between whatever gzstool uses and golang
				// game still accepts it
				{
					name: "compressed",
					fields: fields{
						Flags:        3150304,
						FileCount:    1,
						UnknownCount: 0,
						Version:      1,
						EntriesFromDisk: []EntryFromDisk{{
							BaseDir:    "compressed_dat",
							Path:       "test.lua",
							Compressed: true,
							MetaFlag:   false,
						}},
					},
					wantErr: false,
				},
		*/
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &Qar{
				Magic:        PCMagic,
				Flags:        tt.fields.Flags,
				FileCount:    tt.fields.FileCount,
				UnknownCount: tt.fields.UnknownCount,
				Version:      tt.fields.Version,
				isConsole:    tt.fields.IsConsole,
				Entries:      []Entry{},
			}

			if tt.fields.IsConsole {
				q.Magic = ConsoleMagic
			}

			// add qar entries
			var err error
			for _, v := range tt.fields.EntriesFromDisk {
				e := Entry{
					Header: EntryHeader{
						FilePath:   v.Path,
						Compressed: v.Compressed,
						MetaFlag:   v.MetaFlag,
					},
					DataHeader: DataHeader{
						EncryptionMagic: 0,
						Key:             v.Key,
					},
					Data: nil,
				}

				if v.Encrypted {
					if v.Key == 0 {
						e.DataHeader.Key = 0xCAFEBABE
					} else {
						e.DataHeader.Key = v.Key
					}
				}

				p := filepath.Join(dataDir, v.BaseDir, v.Path)

				if e.Data, err = os.ReadFile(p); err != nil {
					t.Errorf("%s", err.Error())
				}
				q.Entries = append(q.Entries, e)
			}

			out := &util.ByteArrayReaderWriter{}
			if err = q.Write(out, "", false); (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %s, wantErr %v", err.Error(), tt.wantErr)
			}

			diskPath := filepath.Join(dataDir, tt.name+".dat")
			diskData, err := os.ReadFile(diskPath)
			if err != nil {
				t.Fatalf("cannot read %s: %s", diskPath, err.Error())
			}

			if bytes.Compare(diskData, out.Bytes()) != 0 {
				t.Fatalf("not equal\nhave: % x\nwant: % x\n", out.Bytes(), diskData)
			}
		})
	}
}
