package hashing

import "testing"

func TestStrCode64(t *testing.T) {
	type args struct {
		s []byte
	}
	tests := []struct {
		name string
		args args
		want uint64
	}{
		{
			name: "",
			args: args{
				s: []byte("name"),
			},
			want: 0x3391ED17A03A,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StrCode64(tt.args.s); got != tt.want {
				t.Errorf("StrCode64() = 0x%x, want 0x%x", got, tt.want)
			}
		})
	}
}
