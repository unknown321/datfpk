package hashing

import "testing"

func TestStrCode32(t *testing.T) {
	type args struct {
		s []byte
	}
	tests := []struct {
		name string
		args args
		want uint64
	}{
		{
			name: "1",
			args: args{s: []byte("CNP_EYE")},
			want: 0x287a3f51197b9,
		},
		{
			name: "2",
			args: args{s: []byte("SKL_013_LHAND")},
			want: 0xaa0cb6e7389f,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StrCode32(tt.args.s); got != tt.want {
				t.Errorf("StrCode32() = %x, want %x (%s)", got, tt.want, tt.args.s)
			}
		})
	}
}
