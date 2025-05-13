package qar

import (
	"bytes"
	"reflect"
	"testing"
)

func TestDecrypt2Stream_Decrypt2(t *testing.T) {
	type args struct {
		input []byte
		key   uint32
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "reversible",
			args: args{
				input: []byte("123456789123456789123456789"),
				key:   0xCAFEBABE,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Decrypt2Stream{}
			d.Init(tt.args.key)

			got, err := d.Decrypt2(tt.args.input, len(tt.args.input))
			if err != nil {
				t.Errorf("Decrypt2() error = %s", err.Error())
				return
			}

			d1 := &Decrypt2Stream{}
			d1.Init(tt.args.key)
			back, err := d1.Decrypt2(got, len(got))

			if !reflect.DeepEqual(tt.args.input, back) {
				t.Errorf("not equal, have %s, want %s", back, tt.args.input)
			}
		})
	}
}

func TestDecrypt2Stream_Read(t *testing.T) {
	type args struct {
		data []byte
		key  uint32
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "",
			args: args{
				data: []byte("data1234567890\ndata1234567\n"),
				key:  0xcafebabe,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Decrypt2Stream{}
			d.Init(tt.args.key)
			reader := bytes.NewReader(tt.args.data)
			got, err := d.Read(reader, len(tt.args.data))
			if err != nil {
				t.Errorf("%s", err.Error())
				return
			}

			d1 := Decrypt2Stream{}
			d1.Init(tt.args.key)
			gotR := bytes.NewReader(got)
			res, err := d1.Read(gotR, gotR.Len())
			if err != nil {
				t.Errorf("%s", err.Error())
				return
			}
			if bytes.Compare(res, tt.args.data) != 0 {
				t.Errorf("Read() got = %v, want %v", res, tt.args.data)
			}
		})
	}
}
