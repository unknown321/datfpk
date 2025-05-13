package qar

import (
	"bytes"
	"crypto/md5"
	"reflect"
	"testing"
)

func TestDecrypt1Stream_Read(t *testing.T) {
	type args struct {
		data     []byte
		pathHash uint64
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "",
			args: args{
				data:     []byte("some data for testing"),
				pathHash: 0x18e3189a8efa181d,
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Decrypt1Stream{}
			m := Md5Sum{}
			reader := bytes.NewReader(tt.args.data)
			for i, v := range md5.Sum(tt.args.data) {
				m[i] = v
			}
			d.Init(m, tt.args.pathHash, 1, len(tt.args.data))
			got, err := d.Read(reader, len(tt.args.data))
			if err != nil {
				t.Errorf("%s", err.Error())
			}

			gotR := bytes.NewReader(got)
			d1 := Decrypt1Stream{}
			d1.Init(m, tt.args.pathHash, 1, len(tt.args.data))
			reverse, err := d1.Read(gotR, len(tt.args.data))
			if err != nil {
				t.Errorf("%s", err.Error())
			}

			if !reflect.DeepEqual(reverse, tt.args.data) {
				t.Errorf("Read() got = %v, want %v", reverse, tt.args.data)
			}
		})
	}
}
