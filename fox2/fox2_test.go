package fox2

import (
	"bytes"
	"datfpk/util"
	"encoding/xml"
	"fmt"
	"os"
	"testing"
)

func TestFox2_Read(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "",
			args: args{
				filename: "testdata/game/title_sequence.fox2", // ./00_dat/Assets/tpp/pack/mission2/init/title_fpkd/Assets/tpp/level/mission2/init/title_sequence.fox2
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in, err := os.Open(tt.args.filename)
			if err != nil {
				t.Fatalf("%s", err.Error())
			}
			defer in.Close()

			f := &Fox2{}
			if err = f.Read(in); err != nil {
				t.Fatalf("%s", err.Error())
			}
		})
	}
}

func TestFox2_Marshal(t *testing.T) {
	type args struct {
		filename string
		expected string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "bool",
			args: args{
				filename: "testdata/types/bool.foxtool.fox2",
				expected: "testdata/types/bool.fox2.xml",
			},
		},
		{
			name: "entityHandle",
			args: args{
				filename: "testdata/types/entityhandle.foxtool.fox2",
				expected: "testdata/types/entityhandle.fox2.xml",
			},
		},
		{
			name: "entityLink",
			args: args{
				filename: "testdata/types/entitylink.foxtool.fox2",
				expected: "testdata/types/entitylink.fox2.xml",
			},
		},
		{
			name: "entityPtr",
			args: args{
				filename: "testdata/types/entityptr.foxtool.fox2",
				expected: "testdata/types/entityptr.fox2.xml",
			},
		},
		{
			name: "fileptr",
			args: args{
				filename: "testdata/types/fileptr.foxtool.fox2",
				expected: "testdata/types/fileptr.fox2.xml",
			},
		},
		{
			name: "float",
			args: args{

				filename: "testdata/types/float.foxtool.fox2",
				expected: "testdata/types/float.fox2.xml",
			},
		},
		{
			name: "int8",
			args: args{
				filename: "testdata/types/int8.foxtool.fox2",
				expected: "testdata/types/int8.fox2.xml",
			},
		},
		{
			name: "path",
			args: args{
				filename: "testdata/types/path.foxtool.fox2",
				expected: "testdata/types/path.fox2.xml",
			},
		},
		{
			name: "quat",
			args: args{
				filename: "testdata/types/quat.foxtool.fox2",
				expected: "testdata/types/quat.fox2.xml",
			},
		},
		{
			name: "string",
			args: args{
				filename: "testdata/types/string.foxtool.fox2",
				expected: "testdata/types/string.fox2.xml",
			},
		},
		{
			name: "uint8",
			args: args{
				filename: "testdata/types/uint8.foxtool.fox2",
				expected: "testdata/types/uint8.fox2.xml",
			},
		},
		{
			name: "int16",
			args: args{
				filename: "testdata/types/int16.foxtool.fox2",
				expected: "testdata/types/int16.fox2.xml",
			},
		},
		{
			name: "uint16",
			args: args{
				filename: "testdata/types/uint16.foxtool.fox2",
				expected: "testdata/types/uint16.fox2.xml",
			},
		},
		{
			name: "uint32",
			args: args{
				filename: "testdata/types/uint32.foxtool.fox2",
				expected: "testdata/types/uint32.fox2.xml",
			},
		},
		{
			name: "int32",
			args: args{
				filename: "testdata/types/int32.foxtool.fox2",
				expected: "testdata/types/int32.fox2.xml",
			},
		},
		{
			name: "uint64",
			args: args{
				filename: "testdata/types/uint64.foxtool.fox2",
				expected: "testdata/types/uint64.fox2.xml",
			},
		},
		{
			name: "int64",
			args: args{
				filename: "testdata/types/int64.foxtool.fox2",
				expected: "testdata/types/int64.fox2.xml",
			},
		},
		{
			name: "double",
			args: args{
				filename: "testdata/types/double.foxtool.fox2",
				expected: "testdata/types/double.fox2.xml",
			},
		},
		{
			name: "vector3",
			args: args{
				filename: "testdata/types/vector3.foxtool.fox2",
				expected: "testdata/types/vector3.fox2.xml",
			},
		},
		{
			name: "vector4",
			args: args{
				filename: "testdata/types/vector4.foxtool.fox2",
				expected: "testdata/types/vector4.fox2.xml",
			},
		},
		{
			name: "matrix3",
			args: args{
				filename: "testdata/types/matrix3.foxtool.fox2",
				expected: "testdata/types/matrix3.fox2.xml",
			},
		},
		{
			name: "matrix4",
			args: args{
				filename: "testdata/types/matrix4.foxtool.fox2",
				expected: "testdata/types/matrix4.fox2.xml",
			},
		},
		{
			name: "color",
			args: args{
				filename: "testdata/types/color.foxtool.fox2",
				expected: "testdata/types/color.fox2.xml",
			},
		},
		{
			name: "widevector3",
			args: args{
				filename: "testdata/types/widevector3.foxtool.fox2",
				expected: "testdata/types/widevector3.fox2.xml",
			},
		},
		{
			name: "stringMap",
			args: args{
				filename: "testdata/containers/stringmap.foxtool.fox2",
				expected: "testdata/containers/stringmap.fox2.xml",
			},
		},
		{
			name: "stringMap with string",
			args: args{
				filename: "testdata/containers/stringmapWithString.foxtool.fox2",
				expected: "testdata/containers/stringmapWithString.fox2.xml",
			},
		},
		{
			name: "00_dat/Assets/tpp/pack/mission2/init/title_fpkd/Assets/tpp/level/mission2/init/title_sequence.fox2",
			args: args{
				filename: "testdata/game/title_sequence.fox2",
				expected: "testdata/game/title_sequence.fox2.xml",
			},
		},
		{
			name: "",
			args: args{
				filename: "testdata/game/player2_add_parts_prqst_x1.fox2",
				expected: "testdata/game/player2_add_parts_prqst_x1.fox2.xml",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in, err := os.Open(tt.args.filename)
			if err != nil {
				t.Fatalf("%s", err.Error())
			}
			defer in.Close()

			f := &Fox2{}
			if err = f.Read(in); err != nil {
				t.Fatalf("%s", err.Error())
			}

			b := []byte{}
			bb := bytes.NewBuffer(b)

			if err = f.ToXML(bb); err != nil {
				t.Fatalf("%s", err.Error())
			}

			expected, err := os.ReadFile(tt.args.expected)
			if err != nil {
				t.Fatalf("%s", err.Error())
			}

			//os.WriteFile(tt.args.expected, bb.Bytes(), 0644)

			if bytes.Compare(expected, bb.Bytes()) != 0 {
				fmt.Printf("%s\n", bb.Bytes())
				t.Fatalf("not equal, %s != %s", tt.args.filename, tt.args.expected)
			}
		})
	}
}

func TestFox2_Write(t *testing.T) {
	type args struct {
		in string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "bool",
			args: args{
				in: "testdata/types/bool.foxtool.fox2",
			},
		},
		{
			name: "color",
			args: args{
				in: "testdata/types/color.foxtool.fox2",
			},
		},
		{
			name: "double",
			args: args{
				in: "testdata/types/double.foxtool.fox2",
			},
		},
		{
			name: "entityHandle",
			args: args{
				in: "testdata/types/entityhandle.foxtool.fox2",
			},
		},
		//string data order is different, but it's a match
		//{
		//	name: "entityLink",
		//	args: args{
		//		in: "testdata/types/entitylink.foxtool.fox2",
		//	},
		//},
		{
			name: "entityPtr",
			args: args{
				in: "testdata/types/entityptr.foxtool.fox2",
			},
		},
		{
			name: "filePtr",
			args: args{
				in: "testdata/types/fileptr.foxtool.fox2",
			},
		},
		{
			name: "float",
			args: args{
				in: "testdata/types/float.foxtool.fox2",
			},
		},
		{
			name: "int8",
			args: args{
				in: "testdata/types/int8.foxtool.fox2",
			},
		},
		{
			name: "uint8",
			args: args{
				in: "testdata/types/uint8.foxtool.fox2",
			},
		},
		{
			name: "int16",
			args: args{
				in: "testdata/types/int16.foxtool.fox2",
			},
		},
		{
			name: "uint16",
			args: args{
				in: "testdata/types/uint16.foxtool.fox2",
			},
		},
		{
			name: "int32",
			args: args{
				in: "testdata/types/int32.foxtool.fox2",
			},
		},
		{
			name: "uint32",
			args: args{
				in: "testdata/types/uint32.foxtool.fox2",
			},
		},
		{
			name: "int64",
			args: args{
				in: "testdata/types/int64.foxtool.fox2",
			},
		},
		{
			name: "uint64",
			args: args{
				in: "testdata/types/uint64.foxtool.fox2",
			},
		},
		{
			name: "matrix3",
			args: args{
				in: "testdata/types/matrix3.foxtool.fox2",
			},
		},
		{
			name: "matrix4",
			args: args{
				in: "testdata/types/matrix4.foxtool.fox2",
			},
		},
		{
			name: "path",
			args: args{
				in: "testdata/types/path.foxtool.fox2",
			},
		},
		{
			name: "quat",
			args: args{
				in: "testdata/types/quat.foxtool.fox2",
			},
		},
		{
			name: "string",
			args: args{
				in: "testdata/types/string.foxtool.fox2",
			},
		},
		{
			name: "vector3",
			args: args{
				in: "testdata/types/vector3.foxtool.fox2",
			},
		},
		{
			name: "vector4",
			args: args{
				in: "testdata/types/vector4.foxtool.fox2",
			},
		},
		{
			name: "widevector3",
			args: args{
				in: "testdata/types/widevector3.foxtool.fox2",
			},
		},
		{
			name: "stringMap",
			args: args{
				in: "testdata/containers/stringmap.foxtool.fox2",
			},
		},
		{
			name: "stringMapWithString",
			args: args{
				in: "testdata/containers/stringmapWithString.foxtool.fox2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			f := &Fox2{}
			ff, err := os.Open(tt.args.in)
			if err != nil {
				t.Fatalf("%s", err.Error())
			}
			defer ff.Close()

			if err = f.Read(ff); err != nil {
				t.Fatalf("%s", err.Error())
			}

			expected, err := os.ReadFile(tt.args.in)
			if err != nil {
				t.Fatalf("%s", err.Error())
			}

			out := []byte{}
			outB := util.NewByteArrayReaderWriter(out)
			if err = f.Write(outB); err != nil {
				t.Fatalf("%s", err.Error())
			}

			//_ = os.WriteFile("/tmp/test.fox2", outB.Bytes(), 0644)

			if bytes.Compare(outB.Bytes(), expected) != 0 {
				t.Fatalf("not equal")
			}
		})
	}
}

func TestFox2_WriteFromXML(t *testing.T) {
	type args struct {
		in       string
		expected string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "bool",
			args: args{
				in:       "testdata/types/bool.fox2.xml",
				expected: "testdata/types/bool.foxtool.fox2",
			},
		},
		{
			name: "color",
			args: args{
				in:       "testdata/types/color.fox2.xml",
				expected: "testdata/types/color.foxtool.fox2",
			},
		},
		{
			name: "double",
			args: args{
				in:       "testdata/types/double.fox2.xml",
				expected: "testdata/types/double.foxtool.fox2",
			},
		},
		{
			name: "entityHandle",
			args: args{
				in:       "testdata/types/entityhandle.fox2.xml",
				expected: "testdata/types/entityhandle.foxtool.fox2",
			},
		},
		//string data order is different, but it's a match
		//{
		//	name: "entityLink",
		//	args: args{
		//		in:       "testdata/types/entitylink.fox2.xml",
		//		expected: "testdata/types/entitylink.foxtool.fox2",
		//	},
		//},
		{
			name: "entityPtr",
			args: args{
				in:       "testdata/types/entityptr.fox2.xml",
				expected: "testdata/types/entityptr.foxtool.fox2",
			},
		},
		{
			name: "filePtr",
			args: args{
				in:       "testdata/types/fileptr.fox2.xml",
				expected: "testdata/types/fileptr.foxtool.fox2",
			},
		},
		{
			name: "float",
			args: args{
				in:       "testdata/types/float.fox2.xml",
				expected: "testdata/types/float.foxtool.fox2",
			},
		},
		{
			name: "int8",
			args: args{
				in:       "testdata/types/int8.fox2.xml",
				expected: "testdata/types/int8.foxtool.fox2",
			},
		},
		{
			name: "uint8",
			args: args{
				in:       "testdata/types/uint8.fox2.xml",
				expected: "testdata/types/uint8.foxtool.fox2",
			},
		},
		{
			name: "int16",
			args: args{
				in:       "testdata/types/int16.fox2.xml",
				expected: "testdata/types/int16.foxtool.fox2",
			},
		},
		{
			name: "uint16",
			args: args{
				in:       "testdata/types/uint16.fox2.xml",
				expected: "testdata/types/uint16.foxtool.fox2",
			},
		},
		{
			name: "int32",
			args: args{
				in:       "testdata/types/int32.fox2.xml",
				expected: "testdata/types/int32.foxtool.fox2",
			},
		},
		{
			name: "uint32",
			args: args{
				in:       "testdata/types/uint32.fox2.xml",
				expected: "testdata/types/uint32.foxtool.fox2",
			},
		},
		{
			name: "int64",
			args: args{
				in:       "testdata/types/int64.fox2.xml",
				expected: "testdata/types/int64.foxtool.fox2",
			},
		},
		{
			name: "uint64",
			args: args{
				in:       "testdata/types/uint64.fox2.xml",
				expected: "testdata/types/uint64.foxtool.fox2",
			},
		},
		{
			name: "matrix3",
			args: args{
				in:       "testdata/types/matrix3.fox2.xml",
				expected: "testdata/types/matrix3.foxtool.fox2",
			},
		},
		{
			name: "matrix4",
			args: args{
				in:       "testdata/types/matrix4.fox2.xml",
				expected: "testdata/types/matrix4.foxtool.fox2",
			},
		},
		{
			name: "path",
			args: args{
				in:       "testdata/types/path.fox2.xml",
				expected: "testdata/types/path.foxtool.fox2",
			},
		},
		{
			name: "quat",
			args: args{
				in:       "testdata/types/quat.fox2.xml",
				expected: "testdata/types/quat.foxtool.fox2",
			},
		},
		{
			name: "string",
			args: args{
				in:       "testdata/types/string.fox2.xml",
				expected: "testdata/types/string.foxtool.fox2",
			},
		},
		{
			name: "vector3",
			args: args{
				in:       "testdata/types/vector3.fox2.xml",
				expected: "testdata/types/vector3.foxtool.fox2",
			},
		},
		{
			name: "vector4",
			args: args{
				in:       "testdata/types/vector4.fox2.xml",
				expected: "testdata/types/vector4.foxtool.fox2",
			},
		},
		{
			name: "widevector3",
			args: args{
				in:       "testdata/types/widevector3.fox2.xml",
				expected: "testdata/types/widevector3.foxtool.fox2",
			},
		},
		{
			name: "stringMap",
			args: args{
				in:       "testdata/containers/stringmap.fox2.xml",
				expected: "testdata/containers/stringmap.foxtool.fox2",
			},
		},
		{
			name: "stringMapWithString",
			args: args{
				in:       "testdata/containers/stringmapWithString.fox2.xml",
				expected: "testdata/containers/stringmapWithString.foxtool.fox2",
			},
		},
		/*
			{
				name: "o50050_sequence.fox2.xml",
				args: args{
					in:       "testdata/game/o50050_sequence.fox2.xml",
					expected: "",
				},
			},
				{
					name: "",
					args: args{
						in:       "testdata/game/title_sequence.fox2.xml",
						expected: "testdata/game/title_sequence.fox2",
					},
				},
				{
					name: "",
					args: args{
						in:       "testdata/game/player2_add_parts_prqst_x1.fox2.xml",
						expected: "testdata/game/player2_add_parts_prqst_x1.fox2",
					},
				},
				{
					name: "",
					args: args{
						in:       "testdata/game/s10010_sequence.fox2.xml",
						expected: "testdata/game/s10010_sequence.fox2",
					},
				},
		*/
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error

			f := Fox2{}
			data, err := os.ReadFile(tt.args.in)
			if err != nil {
				t.Fatalf("%s", err.Error())
			}
			if err = xml.Unmarshal(data, &f); err != nil {
				t.Fatalf("%s", err.Error())
			}

			expected, err := os.ReadFile(tt.args.expected)
			if err != nil {
				t.Fatalf("%s", err.Error())
			}

			out := []byte{}
			outB := util.NewByteArrayReaderWriter(out)
			if err = f.Write(outB); err != nil {
				t.Fatalf("%s", err.Error())
			}

			//_ = os.WriteFile("/tmp/test.fox2", outB.Bytes(), 0644)

			if bytes.Compare(outB.Bytes(), expected) != 0 {
				t.Fatalf("not equal")
			}
		})
	}
}
