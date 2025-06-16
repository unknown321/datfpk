package fpk

import (
	"bytes"
	"encoding/json"
	"github.com/r3labs/diff/v3"
	"github.com/unknown321/datfpk/util"
	"os"
	"reflect"
	"testing"
)

const datadir = "testdata/"

func TestHeader_Read(t *testing.T) {
	type args struct {
		filename string
	}
	type want struct {
		IsFpkd       bool
		MagicNumber2 uint32
		FileSize     uint32
		FileCount    uint32
		RefCount     uint32
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "fpkd",
			args: args{
				filename: "o50050_subtitles.fpkd",
			},
			want: want{
				IsFpkd:       true,
				MagicNumber2: 2,
				FileSize:     1984,
				FileCount:    1,
				RefCount:     0,
			},
		},
		{
			name: "fpk",
			args: args{
				filename: "plfova_cmf0_main0_def_v00.fpk",
			},
			want: want{
				IsFpkd:       false,
				MagicNumber2: 2,
				FileSize:     240,
				FileCount:    1,
				RefCount:     0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Header{}
			f, err := os.Open(datadir + tt.args.filename)
			if err != nil {
				t.Fatalf("%s", err.Error())
			}
			defer f.Close()

			if err = h.Read(f); err != nil {
				t.Fatalf("%s", err.Error())
			}

			if h.IsFpkd() != tt.want.IsFpkd {
				t.Fatalf("want %t, got %t", tt.want.IsFpkd, h.IsFpkd())
			}

			if h.MagicNumber2 != tt.want.MagicNumber2 {
				t.Fatalf("magic2, want %d, got %d", tt.want.MagicNumber2, h.MagicNumber2)
			}

			if h.FileSize != tt.want.FileSize {
				t.Fatalf("filesize, want %d, got %d", tt.want.FileSize, h.FileSize)
			}

			if h.RefCount != tt.want.RefCount {
				t.Fatalf("refcount, want %d, got %d", tt.want.RefCount, h.RefCount)
			}

			if h.EntryCount != tt.want.FileCount {
				t.Fatalf("filecount, want %d, got %d", tt.want.FileCount, h.EntryCount)
			}
		})
	}
}

func TestFpk_Read(t *testing.T) {
	type want struct {
		Header  Header
		Entries []Entry
	}
	type args struct {
		filename string
		expected []string
	}
	tests := []struct {
		name string
		want want
		args args
	}{
		{
			name: "fpkd",
			want: want{
				Header: Header{
					Magic:        [10]byte{},
					FileSize:     1984,
					MagicNumber2: 2,
					EntryCount:   1,
					RefCount:     0,
				},
				Entries: []Entry{{
					DataOffset: 176,
					DataSize:   1808,
					FilePath: String{
						Header: StringHeader{
							Offset: 96,
							Skip1:  0,
							Length: 71,
							Skip2:  0,
						},
						Data: "/Assets/tpp/ui/Subtitles/package/EngVoice/EngText/o50050_subtitles.fox2",
					},
					PathMD5: [16]byte{0x16, 0x30, 0xa2, 0xe5, 0x46, 0x93, 0xa2, 0x27, 0x27, 0x46, 0x5a, 0xda, 0x51, 0xf9, 0xdd, 0xbc},
				}},
			},
			args: args{
				filename: "o50050_subtitles.fpkd",
				expected: []string{"o50050_subtitles.fox2"},
			},
		},
		{
			name: "many entries",
			want: want{
				Header: Header{
					Magic:        [10]byte{},
					FileSize:     5424,
					MagicNumber2: 2,
					EntryCount:   3,
					RefCount:     0,
				},
				Entries: []Entry{
					{
						DataOffset: 384,
						DataSize:   656,
						FilePath: String{
							Header: StringHeader{
								Offset: 192,
								Skip1:  0,
								Length: 72,
								Skip2:  0,
							},
							Data: "/Assets/tpp/level_asset/weapon/keep_in_fpkl/EQP_WP_SP_SLD_BASE_keep.fox2",
						},
						PathMD5: [16]byte{0xc6, 0xb6, 0x26, 0xd6, 0x22, 0x49, 0xd5, 0x08, 0x95, 0x59, 0xf7, 0x32, 0x96, 0x81, 0x4c, 0x6f},
						Data:    nil,
					},
					{
						DataOffset: 1040,
						DataSize:   2368,
						FilePath: String{
							Header: StringHeader{
								Offset: 265,
								Skip1:  0,
								Length: 53,
								Skip2:  0,
							},
							Data: "/Assets/tpp/parts/weapon/sld/sd02_main0_def_v00.parts",
						},
						PathMD5: [16]byte{0x8f, 0xc3, 0xee, 0x3d, 0x45, 0xea, 0xf9, 0x41, 0x25, 0xed, 0x22, 0x10, 0x16, 0x1e, 0xe6, 0xfc},
						Data:    nil,
					},
					{
						DataOffset: 3408,
						DataSize:   2016,
						FilePath: String{
							Header: StringHeader{
								Offset: 319,
								Skip1:  0,
								Length: 59,
								Skip2:  0,
							},
							Data: "/Assets/tpp/level_asset/weapon/PhysicsParameter/shield.phsd",
						},
						PathMD5: [16]byte{0xfb, 0xb6, 0xe0, 0x04, 0x47, 0xa2, 0xcf, 0x32, 0x3a, 0x14, 0xab, 0x63, 0xa8, 0x7c, 0x2f, 0xd2},
						Data:    nil,
					}},
			},
			args: args{
				filename: "EQP_WP_SP_SLD_BASE.fpkd",
				expected: []string{"fpkd/EQP_WP_SP_SLD_BASE_keep.fox2", "fpkd/sd02_main0_def_v00.parts", "fpkd/shield.phsd"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Fpk{}
			in, err := os.Open(datadir + tt.args.filename)
			if err != nil {
				t.Fatalf("%s", err.Error())
			}
			defer in.Close()

			if err = f.Read(in, false); err != nil {
				t.Fatalf("%s", err.Error())
			}

			magic := MagicFpk
			if f.Header.IsFpkd() {
				magic = MagicFpkd
			}
			for k, v := range magic {
				tt.want.Header.Magic[k] = v
			}

			for k := range tt.want.Entries {
				tt.want.Entries[k].Data, err = os.ReadFile(datadir + tt.args.expected[k])

				if err != nil {
					t.Fatalf("%s", err)
				}
			}

			if !reflect.DeepEqual(f.Header, tt.want.Header) {
				d, err := diff.Diff(f.Header, tt.want.Header)
				if err != nil {
					t.Fatalf("%s", err.Error())
				}

				b, err := json.Marshal(d)
				if err != nil {
					t.Fatalf("%s", err.Error())
				}

				t.Fatalf("%s", b)
			}

			if !reflect.DeepEqual(f.Entries, tt.want.Entries) {
				d, err := diff.Diff(f.Entries, tt.want.Entries)
				if err != nil {
					t.Fatalf("%s", err.Error())
				}

				b, err := json.Marshal(d)
				if err != nil {
					t.Fatalf("%s", err.Error())
				}

				t.Fatalf("%s", b)
			}
		})
	}
}

func TestFpk_ReadReference(t *testing.T) {
	in, err := os.Open("testdata/wfv_camo_c45.fpk")
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	defer in.Close()

	f := Fpk{}
	if err = f.Read(in, false); err != nil {
		t.Fatalf("%s", err.Error())
	}

	expected := Fpk{
		Header: Header{
			Magic:        [10]byte{},
			FileSize:     464,
			MagicNumber2: 2,
			EntryCount:   1,
			RefCount:     2,
		},
		Entries: []Entry{},
		References: []Reference{
			{
				FilePath: String{
					Header: StringHeader{
						Offset: 173,
						Skip1:  0,
						Length: 54,
						Skip2:  0,
					},
					Data: "/Assets/tpp/pack/collectible/common/col_common_tpp.fpk",
				},
			},
			{
				FilePath: String{
					Header: StringHeader{
						Offset: 228,
						Skip1:  0,
						Length: 40,
						Skip2:  0,
					},
					Data: "/Assets/tpp/pack/resident/resident00.fpk",
				},
			}},
	}

	for v := range MagicFpk {
		expected.Header.Magic[v] = MagicFpk[v]
	}

	if !reflect.DeepEqual(expected.Header, f.Header) {
		t.Fatalf("header not equal, have %+v, want %+v", f.Header, expected.Header)
	}

	if !reflect.DeepEqual(expected.References, f.References) {
		t.Fatalf("references not equal, have %+v, want %+v", f.References, expected.References)
	}
}

func TestFpk_Write(t *testing.T) {
	baseDir := "testdata/wfv_camo_c45_fpk"
	files := []string{"/Assets/tpp/fova/weapon/all/wfv_camo_c45.fv2"}
	refs := []string{"/Assets/tpp/pack/collectible/common/col_common_tpp.fpk", "/Assets/tpp/pack/resident/resident00.fpk"}

	f := Fpk{
		Header: Header{
			Magic: MagicFpk,
		},
		Entries:    []Entry{},
		References: []Reference{},
	}

	for _, r := range refs {
		f.References = append(f.References, Reference{FilePath: String{Data: r}})
	}

	for _, e := range files {
		entry := Entry{
			FilePath: String{
				Data: e,
			},
			Data:      []byte{},
			Encrypted: false,
		}

		f.Entries = append(f.Entries, entry)
	}

	b := &util.ByteArrayReaderWriter{}
	if err := f.Write(b, baseDir, false); err != nil {
		t.Fatalf("%s", err.Error())
	}

	expected, err := os.ReadFile("testdata/wfv_camo_c45.fpk")
	if err != nil {
		t.Fatalf("%s", err.Error())
	}

	if bytes.Compare(b.Bytes(), expected) != 0 {
		t.Fatalf("not equal")
	}

	//rr, err := os.OpenFile("/tmp/out.fpk", os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0644)
	//if err != nil {
	//	t.Fatalf("%s", err.Error())
	//}
	//rr.Write(b.Bytes())
	//rr.Close()
}
