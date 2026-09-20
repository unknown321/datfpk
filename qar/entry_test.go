package qar

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/unknown321/datfpk/crypto"
	"github.com/unknown321/datfpk/util"
	"github.com/unknown321/hashing"
)

func TestMd5Sum_UnmarshalJSON(t *testing.T) {
	j := `{
            "filePath": "/Fox/Scripts/Gr/gr_init_dx11.lua",
            "metaFlag": true,
            "compressed": false,
            "encryption": 2700668902,
			"key": 2785280071
        }`

	e := &Entry{}
	if err := json.Unmarshal([]byte(j), e); err != nil {
		t.Errorf("fail: %s", err.Error())
	}

	if e.Header.FilePath != "/Fox/Scripts/Gr/gr_init_dx11.lua" {
		t.Errorf("fp: %s", e.Header.FilePath)
	}

	if e.DataHeader.EncryptionMagic != 0xA0F8EFE6 {
		t.Errorf("encryption: %d", e.DataHeader.EncryptionMagic)
	}

	if e.DataHeader.Key != 2785280071 {
		t.Errorf("encryption: %d", e.DataHeader.EncryptionMagic)
	}
}

func TestEntry_MarshalJSON(t *testing.T) {
	type fields struct {
		DataHeader DataHeader
		Header     EntryHeader
		Data       []byte
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			name: "1",
			fields: fields{
				DataHeader: DataHeader{
					EncryptionMagic: 1,
					Key:             2,
				},
				Header: EntryHeader{
					PathHash:         0,
					UncompressedSize: 0,
					CompressedSize:   0,
					Md5Sum:           Md5Sum{},
					FilePath:         "/test123",
					DataOffset:       0,
					Compressed:       false,
					Version:          0,
					MetaFlag:         false,
				},
				Data: nil,
			},
			want:    []byte(`{"filePath":"/test123","encryption":1,"key":2}`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Entry{
				DataHeader: tt.fields.DataHeader,
				Header:     tt.fields.Header,
				Data:       tt.fields.Data,
			}
			got, err := e.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MarshalJSON() got = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestEntry_Read(t *testing.T) {
	//dd if=00.dat bs=1 skip=496323584 count=500 of=foxdat
	expected, _ := os.ReadFile("testdata/foxpatch.dat")

	f, err := os.Open("testdata/foxdat")
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	defer f.Close()

	e := Entry{}
	if err = e.Read(f, 1, binary.LittleEndian); err != nil {
		t.Fatalf("%s", err.Error())
	}
	if err = e.ReadData(f); err != nil {
		t.Fatalf("%s", err)
	}
	if bytes.Compare(e.Data, expected) != 0 {
		t.Fatalf("not equal")
	}
}

func TestEntry_ReadCompressed(t *testing.T) {
	// dd if=00.dat bs=1 skip=324063232 count=400 of=plfova_cmf0_main0_def_v00.fpk.compressed
	var err error
	expected, err := os.ReadFile("testdata/plfova_cmf0_main0_def_v00.fpk")
	if err != nil {
		t.Fatalf("%s", err.Error())
	}

	f, err := os.Open("testdata/plfova_cmf0_main0_def_v00.fpk.compressed")
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	defer f.Close()

	e := Entry{}
	if err = e.Read(f, 1, binary.LittleEndian); err != nil {
		t.Fatalf("%s", err.Error())
	}
	if err = e.ReadData(f); err != nil {
		t.Fatalf("%s", err)
	}
	if bytes.Compare(e.Data, expected) != 0 {
		t.Fatalf("not equal, have %s, want %s", e.Data, expected)
	}
}

func TestEntry_Write(t *testing.T) {
	type fields struct {
		FilePath   string
		Compressed bool
		Version    uint32
		MetaFlag   bool
		Encrypted  bool
		Data       []byte
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "plain",
			fields: fields{
				FilePath:   "test.lua",
				Compressed: false,
				Version:    1,
				MetaFlag:   true,
				Data:       []byte("1234567890987654321\n"),
			},
		},
		{
			name: "compressed",
			fields: fields{
				FilePath:   "test.lua",
				Compressed: true,
				Version:    1,
				MetaFlag:   true,
				Data:       []byte("data1234567890\ndata1234567\n"),
			},
		},
		{
			name: "encrypted",
			fields: fields{
				FilePath:   "foxpatch.dat",
				Compressed: false,
				Version:    1,
				MetaFlag:   true,
				Encrypted:  true,
				Data:       nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error

			in := &Entry{
				Header: EntryHeader{
					FilePath:   tt.fields.FilePath,
					Compressed: tt.fields.Compressed,
					Version:    tt.fields.Version,
					MetaFlag:   tt.fields.MetaFlag,
				},
				DataHeader: DataHeader{
					EncryptionMagic: 0,
					Key:             0,
				},
				Data: tt.fields.Data,
			}

			if tt.fields.Encrypted {
				in.DataHeader.Key = 0xcb830057
				in.Data, err = os.ReadFile("testdata/foxpatch.dat")
			}

			data := []byte{}
			if data, err = in.Write(); err != nil {
				t.Errorf("%s", err)
			}

			out := &Entry{}
			writer := util.NewByteArrayReaderWriter(data)

			if err = out.Read(writer, in.Header.Version, binary.LittleEndian); err != nil {
				t.Errorf("%s", err.Error())
			}

			if err = out.ReadData(writer); err != nil {
				t.Errorf("%s", err.Error())
			}

			if bytes.Compare(out.Data, in.Data) != 0 {
				t.Errorf("not equal\nhave %s\nwant %s", out.Data, in.Data)
			}
		})
	}
}

func TestEntry_WriteConsole(t *testing.T) {
	type fields struct {
		Header     EntryHeader
		DataHeader DataHeader
		Data       []byte
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			name: "not encrypted",
			fields: fields{
				Header: EntryHeader{
					PathHash:           0,
					UncompressedSize:   0,
					CompressedSize:     0,
					Md5Sum:             Md5Sum{},
					FilePath:           "",
					DataOffset:         0,
					Compressed:         false,
					Version:            0,
					MetaFlag:           false,
					NameHashForPacking: 0,
				},
				DataHeader: DataHeader{
					EncryptionMagic:  0,
					Key:              0,
					CompressedSize:   0,
					UncompressedSize: 0,
					Version:          0,
					byteOrder:        nil,
				},
				Data: []byte("return 'data1234567890'\n"),
			},
			want:    []byte("return 'data1234567890'\n"),
			wantErr: false,
		},
		{
			name: "encrypted",
			fields: fields{
				Header: EntryHeader{},
				DataHeader: DataHeader{
					EncryptionMagic:  crypto.Magic2,
					Key:              0x8eb31c0f,
					CompressedSize:   0,
					UncompressedSize: 0,
					Version:          0,
					byteOrder:        binary.LittleEndian,
				},
				Data: []byte("return 'data1234567890'\n"),
			},
			want: []byte{
				0xE6, 0xEF, 0xF8, 0xE3, 0x0F, 0x1C, 0xB3, 0x8E, 0x18, 0x00, 0x00, 0x00, 0x18, 0x00, 0x00, 0x00,
				0x7D, 0x79, 0x8F, 0x8A, 0x4F, 0xED, 0x57, 0x72, 0x97, 0x78, 0x2B, 0x55, 0x20, 0x1B, 0x03, 0xD0,
				0xC2, 0xEF, 0xFC, 0x20, 0xBC, 0x14, 0x11, 0xDE,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Entry{
				Header:     tt.fields.Header,
				DataHeader: tt.fields.DataHeader,
				Data:       tt.fields.Data,
			}
			got, err := e.WriteConsole()
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteConsole() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if bytes.Compare(got, tt.want) != 0 {
				t.Errorf("WriteConsole() = \n% X, want\n% X", got, tt.want)
			}
		})
	}
}

func TestEntry_ReadDataConsole(t *testing.T) {
	type want struct {
		Header     EntryHeader
		DataHeader DataHeader
		Data       []byte
	}
	type input struct {
		f io.ReadSeeker
	}
	tests := []struct {
		name    string
		input   input
		want    want
		wantErr bool
	}{
		{
			name: "not encrypted",
			want: want{
				Header: EntryHeader{
					PathHash:           hashing.HashFileNameWithExtension("test.lua"),
					UncompressedSize:   24,
					CompressedSize:     24,
					Md5Sum:             Md5Sum{},
					FilePath:           "test.lua",
					DataOffset:         0,
					Compressed:         false,
					Version:            2,
					MetaFlag:           false,
					NameHashForPacking: 0,
				},
				DataHeader: DataHeader{
					EncryptionMagic:  0,
					Key:              0,
					CompressedSize:   0,
					UncompressedSize: 0,
					Version:          0,
					byteOrder:        nil,
				},
				Data: []byte("return 'data1234567890'\n"),
			},
			input: input{
				f: bytes.NewReader([]byte("return 'data1234567890'\n")),
			},
			wantErr: false,
		},
		{
			name: "encrypted",
			want: want{
				Header: EntryHeader{
					PathHash:           hashing.HashFileNameWithExtension("test.lua"),
					UncompressedSize:   24 + 16,
					CompressedSize:     24 + 16,
					Md5Sum:             Md5Sum{},
					FilePath:           "test.lua",
					DataOffset:         0,
					Compressed:         false,
					Version:            2,
					MetaFlag:           false,
					NameHashForPacking: 0,
				},
				DataHeader: DataHeader{
					EncryptionMagic:  crypto.Magic2,
					Key:              0x8eb31c0f,
					CompressedSize:   0,
					UncompressedSize: 0,
					Version:          0,
					byteOrder:        nil,
				},
				Data: []byte("return 'data1234567890'\n"),
			},
			input: input{
				f: bytes.NewReader([]byte{
					0xE6, 0xEF, 0xF8, 0xE3, 0x0F, 0x1C, 0xB3, 0x8E, 0x18, 0x00, 0x00, 0x00, 0x18, 0x00, 0x00, 0x00,
					0x7D, 0x79, 0x8F, 0x8A, 0x4F, 0xED, 0x57, 0x72, 0x97, 0x78, 0x2B, 0x55, 0x20, 0x1B, 0x03, 0xD0,
					0xC2, 0xEF, 0xFC, 0x20, 0xBC, 0x14, 0x11, 0xDE,
				}),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Entry{
				Header:     tt.want.Header,
				DataHeader: tt.want.DataHeader,
				Data:       tt.want.Data,
			}

			if err := e.ReadDataConsole(tt.input.f); (err != nil) != tt.wantErr {
				t.Errorf("ReadDataConsole() error = %v, wantErr %v", err, tt.wantErr)
			}

			if bytes.Compare(e.Data, tt.want.Data) != 0 {
				t.Errorf("data not equal\nhave %s\nwant %s", e.Data, tt.want.Data)
			}

			if e.DataHeader.EncryptionMagic != tt.want.DataHeader.EncryptionMagic {
				t.Errorf("encryption magic not equal, have %x, want %x", e.DataHeader.EncryptionMagic, tt.want.DataHeader.EncryptionMagic)
			}

			if e.DataHeader.Key != tt.want.DataHeader.Key {
				t.Errorf("key not equal, have %x, want %x", e.DataHeader.Key, tt.want.DataHeader.Key)
			}
		})
	}
}
