package fox

import "testing"

func TestDataTypeFromString(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    FDataType
		wantErr bool
	}{
		{
			name: "",
			args: args{
				s: "String",
			},
			want:    FString,
			wantErr: false,
		},
		{
			name: "",
			args: args{
				s: "WideVector3",
			},
			want:    FWideVector3,
			wantErr: false,
		},
		{
			name: "",
			args: args{
				s: "asdsad",
			},
			want:    FFail,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DataTypeFromString(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("DataTypeFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DataTypeFromString() got = %v, want %v", got, tt.want)
			}
		})
	}
}
