package qar

import (
	"bytes"
	"datfpk/util"
	"encoding/json"
	"os"
	"reflect"
	"testing"
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
	if err = e.Read(f, 1); err != nil {
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
	if err = e.Read(f, 1); err != nil {
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

			if err = out.Read(writer, in.Header.Version); err != nil {
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
