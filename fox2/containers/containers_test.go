package containers

import "testing"

func TestContainerTypeFromString(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    FoxContainerType
		wantErr bool
	}{
		{
			name: "",
			args: args{
				s: "StringMap",
			},
			want:    StringMap,
			wantErr: false,
		},
		{
			name: "",
			args: args{
				s: "List",
			},
			want:    List,
			wantErr: false,
		},
		{
			name: "",
			args: args{
				s: "fail",
			},
			want:    ContainerTypeUnknown,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ContainerTypeFromString(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ContainerTypeFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ContainerTypeFromString() got = %v, want %v", got, tt.want)
			}
		})
	}
}
