package hashing

import "testing"

func TestHashFileName(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want uint64
	}{
		{
			name: "1",
			args: args{
				s: "/Assets/tpp/ui/Script/emblem_list.lua",
			},
			want: 0x18e393985a3ddd5e,
		},
		{
			name: "2",
			args: args{
				s: "/Assets/tpp/pack/mbdvc/mb_child_window_pool_reward.fpk",
			},
			want: 0x5229cf666bb3452f,
		},
		{
			name: "3",
			args: args{
				s: "foxpatch.dat",
			},
			want: 0xac84c1d94d86e6f1,
		},
		{
			name: "4",
			args: args{
				s: "/Assets/tpp/pack/fova/common_source/chara/cm_head/face/cm_dds3_eqhd6_eye0.fpk",
			},
			want: 0x5228b8101b69f4c5,
		},
		{
			name: "5",
			args: args{
				s: "/init.lua",
			},
			want: 0x18e72be6ec606799,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HashFileNameWithExtension(tt.args.s); got != tt.want {
				t.Errorf("HashFileName() = 0x%x, want 0x%x (%s)", got, tt.want, tt.args.s)
			}
		})
	}
}
