package fpk

import (
	"bytes"
	"os"
	"reflect"
	"testing"
)

func TestDecrypt(t *testing.T) {
	type args struct {
		filename  string
		entryName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "encrypted",
			args: args{
				//dd if=title.fpkd bs=1 skip=7344 count=1008 of=fpkd_encrypted_entry
				filename:  "fpkd_encrypted_entry",
				entryName: "/Assets/tpp/script/mission/mission_main.lua",
			},
			want: "mission_main.lua",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := os.ReadFile("testdata/" + tt.args.filename)
			if err != nil {
				t.Fatalf("%s", err.Error())
			}

			got, err := Decrypt(data, tt.args.entryName)
			if err != nil {
				t.Fatalf("%s", err.Error())
			}

			var want []byte
			want, err = os.ReadFile("testdata/" + tt.want)
			if err != nil {
				t.Fatalf("%s", err.Error())
			}

			if !reflect.DeepEqual(got, want) {
				t.Errorf("Decrypt() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEncrypt(t *testing.T) {
	in, err := os.ReadFile("testdata/mission_main.lua")
	if err != nil {
		t.Fatalf("%s", err.Error())
	}

	name := "/Assets/tpp/script/mission/mission_main.lua"
	res := Encrypt(in, name)
	if res[0] != 0x1B {
		t.Fatalf("missing encryption byte")
	}

	decoded, err := Decrypt(res, name)
	if bytes.Compare(decoded, in) != 0 {
		t.Fatalf("decoded != incoming")
	}

	expected, err := os.ReadFile("testdata/fpkd_encrypted_entry")
	if err != nil {
		t.Fatalf("%s", err.Error())
	}

	if bytes.Compare(expected, res) != 0 {
		t.Fatalf("not equal")
	}
}
