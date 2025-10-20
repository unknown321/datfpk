package lng

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"slices"
	"testing"

	"github.com/unknown321/datfpk/dictionary"
	"github.com/unknown321/datfpk/util"
)

func TestLng_Read(t *testing.T) {
	type fields struct {
		Header  Header
		Entries []Entry
	}
	type args struct {
		filename string
	}
	tests := []struct {
		name     string
		filename string
		fields   fields
		args     args
		wantErr  bool
	}{
		{
			name: "tpp eng",
			fields: fields{
				Header: Header{
					Magic:        Magic,
					Version:      3,
					Endianness:   EndiannessBE,
					EntryCount:   92,
					ValuesOffset: 24,
					KeysOffset:   3692,
				},
				Entries: nil,
			},
			args: args{
				filename: "testdata/tpp_tutorial.eng.lng2",
			},
			wantErr: false,
		},
	}

	dict := dictionary.DictStrCode64{}
	f, err := os.OpenFile("testdata/dict.txt", os.O_RDONLY, 0666)
	if err != nil {
		t.Fatal(err)
	}

	if err = dict.Read(f); err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Lng{
				Header:  tt.fields.Header,
				Entries: tt.fields.Entries,
			}

			seeker, err := os.ReadFile(tt.args.filename)
			if err != nil {
				t.Fatal(err)
			}

			reader := bytes.NewReader(seeker)

			if err := l.Read(reader, dict); (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(l.Header, tt.fields.Header) {
				t.Errorf("Header = %v, want %v", l.Header, tt.fields.Header)
			}
		})
	}
}

func TestLng_MarshalJSON(t *testing.T) {
	tests := []struct {
		name             string
		filename         string
		expectedFilename string
		wantErr          bool
	}{
		{
			name:             "",
			filename:         "testdata/tpp_tutorial.eng.lng2",
			expectedFilename: "testdata/tpp_tutorial.eng.lng2.json",
			wantErr:          false,
		},
		{
			name:             "jpn",
			filename:         "testdata/tpp_tutorial.jpn.lng2",
			expectedFilename: "testdata/tpp_tutorial.jpn.lng2.json",
			wantErr:          false,
		},
	}

	dict := dictionary.DictStrCode64{}
	f, err := os.OpenFile("testdata/dict.txt", os.O_RDONLY, 0666)
	if err != nil {
		t.Fatal(err)
	}

	if err = dict.Read(f); err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Lng{}
			seeker, err := os.ReadFile(tt.filename)
			if err != nil {
				t.Fatal(err)
			}

			reader := bytes.NewReader(seeker)
			if err = l.Read(reader, dict); err != nil {
				t.Fatal(err)
			}

			got, err := l.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			want, err := os.ReadFile(tt.expectedFilename)
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(got, want) {
				t.Errorf("MarshalJSON() got = %s, want %s", got, want)
			}
		})
	}
}

func TestLng_UnmarshalJSON(t *testing.T) {
	type fields struct {
		Header  Header
		Entries []Entry
		Keys    []Key
	}
	type args struct {
		filename string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "eng",
			fields: fields{
				Header: Header{
					Version:    3,
					Endianness: EndiannessBE,
				},
				Entries: []Entry{
					{
						LangId: "tutorial_bino",
						Color:  1,
						Value:  "<I=G=BINOS> Hold: Binoculars",
					},
					{
						LangId: "tutorial_bino_zoom",
						Color:  1,
						Value:  "<I=G=PAD_R3>: Change zoom",
					},
					{
						LangId: "tutorial_optionalradio",
						Color:  1,
						Value:  "<I=G=CALL> Tap: Intel support call",
					},
					{
						LangId: "tutorial_advice",
						Color:  1,
						Value:  "<I=G=CALL> Tap: Radio (request intel)",
					},
					{
						LangId: "tutorial_set_marker",
						Color:  1,
						Value:  "<I=G=HOLD>: Place/remove marker",
					},
					{
						LangId: "",
						Color:  1,
						Value:  "<I=G=HOLD>: Launch",
						Key:    3788606071,
					},
					{
						LangId: "tutorial_smoke",
						Color:  1,
						Value:  "<I=G=HOLD>: Disperse smoke",
					},
					{
						LangId: "",
						Color:  1,
						Value:  "<I=G=HOLD>: Disperse sleeping gas",
						Key:    1193010912,
					},
					{
						LangId: "tutorial_shield",
						Color:  1,
						Value:  "<I=G=PAD_L2>: Drop shield",
					},
					{
						LangId: "",
						Color:  1,
						Value:  "<I=G=PAD_Y>: Change function",
						Key:    4149766436,
					},
					{
						LangId: "tutorial_horse_call",
						Color:  1,
						Value:  "<I=G=PAD_RS> while holding <I=G=CALL>: Call horse",
					},
					{
						LangId: "tutorial_horse_hide",
						Color:  1,
						Value:  "<I=G=STANCE>: Hide on horseback/return upright",
					},
					{
						LangId: "tutorial_horse_hide_change",
						Color:  1,
						Value:  "<I=G=STOCK>: Change hiding position left/right",
					},
					{
						LangId: "tutorial_horse_run",
						Color:  1,
						Value:  "<I=G=EVADE>: Speed up horse",
					},
					{
						LangId: "tutorial_horse_rideon",
						Color:  1,
						Value:  "<I=G=ACTION>: Mount/dismount horse",
					},
					{
						LangId: "tutorial_horse_puton",
						Color:  1,
						Value:  "<I=G=RELOAD> Hold: Place on horse",
					},
					{
						LangId: "tutorial_get_info",
						Color:  1,
						Value:  "<I=G=ACTION>: Acquire intel",
					},
					{
						LangId: "",
						Color:  1,
						Value:  "<I=G=ACTION>: Untie restraints",
						Key:    4168838809,
					},
					{
						LangId: "tutorial_mb_device",
						Color:  1,
						Value:  "<I=G=MBDEVICE>: Open/close iDroid",
					},
					{
						LangId: "tutorial_searchlight_onoff",
						Color:  1,
						Value:  "<I=G=PAD_RIGHT>: Turn light ON/OFF",
					},
					{
						LangId: "tutorial_toilet_hide",
						Color:  1,
						Value:  "<I=G=ACTION>: Hide inside toilet/exit toilet",
					},
					{
						LangId: "tutorial_toilet_putin",
						Color:  1,
						Value:  "<I=G=RELOAD>: Place inside toilet",
					},
					{
						LangId: "tutorial_garbagebox_hide",
						Color:  1,
						Value:  "<I=G=ACTION>: Hide inside dumpster/exit dumpster",
					},
					{
						LangId: "tutorial_garbagebox_putin",
						Color:  1,
						Value:  "<I=G=RELOAD>: Place inside dumpster",
					},
					{
						LangId: "tutorial_ladder",
						Color:  1,
						Value:  "<I=G=ACTION>: Climb up/down ladder",
					},
					{
						LangId: "tutorial_cliff",
						Color:  1,
						Value:  "<I=G=ACTION>: Climb up/down cliff",
					},
					{
						LangId: "tutorial_power_onoff",
						Color:  1,
						Value:  "<I=G=ACTION>: Turn power ON/OFF",
					},
					{
						LangId: "tutorial_attack",
						Color:  1,
						Value:  "<I=G=HOLD>+<I=G=ATTACK>: Use weapon (ready + attack)",
					},
					{
						LangId: "tutorial_elude_up",
						Color:  1,
						Value:  "<I=G=ACTION>: Climb back up",
					},
					{
						LangId: "tutorial_elude_down",
						Color:  1,
						Value:  "<I=G=STANCE>: Drop down",
					},
					{
						LangId: "tutorial_cure",
						Color:  1,
						Value:  "<I=G=ACTION>: First Aid",
					},
					{
						LangId: "tutorial_reload",
						Color:  1,
						Value:  "<I=G=RELOAD>: Reload",
					},
					{
						LangId: "tutorial_jump",
						Color:  1,
						Value:  "<I=G=ACTION> while tilting <I=G=MOVE>: Jump ",
					},
					{
						LangId: "tutorial_shoulder",
						Color:  1,
						Value:  "<I=G=RELOAD> Hold: Carry/set down",
					},
					{
						LangId: "tutorial_shoulder_throw",
						Color:  1,
						Value:  "<I=G=ATTACK>: Throw",
					},
					{
						LangId: "tutorial_stance",
						Color:  1,
						Value:  "<I=G=STANCE>: Change stance (stand/crouch/prone)",
					},
					{
						LangId: "tutorial_equipment_wp",
						Color:  1,
						Value:  "<I=G=PAD_RS> while holding <I=G=PAD_ALL>: Select/equip weapon or item",
					},
					{
						LangId: "tutorial_look_in",
						Color:  1,
						Value:  "<I=G=STOCK> Hold: Peek",
					},
					{
						LangId: "tutorial_fence_jump",
						Color:  1,
						Value:  "<I=G=ACTION>: Climb obstacle/fence",
					},
					{
						LangId: "tutorial_rolling",
						Color:  1,
						Value:  "<I=G=MOVE>+<I=G=DASH> while holding <I=G=HOLD>: Roll (move + roll)",
					},
					{
						LangId: "tutorial_cover_attack",
						Color:  1,
						Value:  "<I=G=HOLD>+<I=G=ATTACK>: Cover attack (ready + attack)",
					},
					{
						LangId: "tutorial_pause",
						Color:  1,
						Value:  "<I=G=PAUSE>: Pause",
					},
					{
						LangId: "tutorial_vehicle_rideon",
						Color:  1,
						Value:  "<I=G=ACTION>: Enter vehicle",
					},
					{
						LangId: "tutorial_vehicle_puton",
						Color:  1,
						Value:  "<I=G=RELOAD> Hold: Place in vehicle",
					},
					{
						LangId: "tutorial_accelarater",
						Color:  1,
						Value:  "<I=G=PAD_R2>: Accelerate",
					},
					{
						LangId: "tutorial_brake",
						Color:  1,
						Value:  "<I=G=PAD_L2>: Brake/reverse",
					},
					{
						LangId: "tutorial_heli_rideon",
						Color:  1,
						Value:  "<I=G=ACTION>: Board helicopter",
					},
					{
						LangId: "tutorial_heli_puton",
						Color:  1,
						Value:  "<I=G=RELOAD> Hold: Place in helicopter",
					},
					{
						LangId: "tutorial_attack_machinegun",
						Color:  1,
						Value:  "<I=G=ATTACK>: Attack using machine gun",
					},
					{
						LangId: "tutorial_attack_mortar",
						Color:  1,
						Value:  "<I=G=ATTACK>: Attack using mortar",
					},
					{
						LangId: "tutorial_attack_anti_aircraft",
						Color:  1,
						Value:  "<I=G=ATTACK>: Attack using anti-air cannon",
					},
					{
						LangId: "tutorial_hulton",
						Color:  1,
						Value:  "<I=G=ACTION> Hold: Fulton recovery",
					},
					{
						LangId: "tutorial_pipe",
						Color:  1,
						Value:  "<I=G=ACTION>: Climb up/down pipe",
					},
					{
						LangId: "tutorial_cqc",
						Color:  1,
						Value:  "<I=G=ATTACK>: CQC (strike/throw/restrain)",
					},
					{
						LangId: "tutorial_cqc_punch",
						Color:  1,
						Value:  "<I=G=ATTACK> Repeatedly: Strike",
					},
					{
						LangId: "tutorial_cqc_throw",
						Color:  1,
						Value:  "Tap <I=G=ATTACK> while tilting <I=G=MOVE>: throw",
					},
					{
						LangId: "tutorial_restraint2",
						Color:  1,
						Value:  "<I=G=ATTACK> Hold: Restrain",
					},
					{
						LangId: "tutorial_restraint",
						Color:  1,
						Value:  "<I=G=CQC> Hold: Restrain",
					},
					{
						LangId: "tutorial_interrogation",
						Color:  1,
						Value:  "<I=G=PAD_RS> while holding <I=G=CALL>: Interrogate",
					},
					{
						LangId: "tutorial_swoon",
						Color:  1,
						Value:  "<I=G=ATTACK> Repeatedly: Choke (knock out)",
					},
					{
						LangId: "tutorial_kill",
						Color:  1,
						Value:  "<I=G=ACTION>: Slit throat (kill)",
					},
					{
						LangId: "",
						Color:  1,
						Value:  "While striking, hold <I=G=HOLD>: Steal weapon",
						Key:    524352877,
					},
					{
						LangId: "tutorial_cqc_combo",
						Color:  1,
						Value:  "After throwing, tap <I=G=ATTACK> again: Consecutive CQC",
					},
					{
						LangId: "tutorial_get_item",
						Color:  1,
						Value:  "<I=G=RELOAD> Hold: Pick up item",
					},
					{
						LangId: "",
						Color:  1,
						Value:  "<I=G=ACTION>: Pick up supply drop",
						Key:    2318475079,
					},
					{
						LangId: "tutorial_C4_set",
						Color:  1,
						Value:  "<I=G=HOLD>+<I=G=ATTACK>: Place C-4 (ready + place)",
					},
					{
						LangId: "tutorial_C4_exploding",
						Color:  1,
						Value:  "<I=G=HOLD>+<I=G=ACTION>: Detonate C-4 (ready + detonate)",
					},
					{
						LangId: "tutorial_change_camera",
						Color:  1,
						Value:  "<I=G=BINOS> Tap: Switch camera (shoulder/first-person)",
					},
					{
						LangId: "tutorial_heli_getoff",
						Color:  1,
						Value:  "<I=G=ACTION>: Disembark helicopter",
					},
					{
						LangId: "tutorial_knock",
						Color:  1,
						Value:  "<I=G=PAD_RS> while holding <I=G=CALL>: Knock (lure enemy)",
					},
					{
						LangId: "tutorial_change_barrel",
						Color:  1,
						Value:  "<I=G=HOLD>+<I=G=ACTION>: Switch to underbarrel (ready + change)",
					},
					{
						LangId: "tutorial_camera_move",
						Color:  1,
						Value:  "<I=G=CAMERA>: Move camera",
					},
					{
						LangId: "tutorial_camera_zoom",
						Color:  1,
						Value:  "<I=G=CAMERA> Hold: Zoom camera",
					},
					{
						LangId: "tutorial_camera_change",
						Color:  1,
						Value:  "<I=G=STOCK> Tap: Switch camera position left/right",
					},
					{
						LangId: "tutorial_play_move",
						Color:  1,
						Value:  "<I=G=MOVE>: Move",
					},
					{
						LangId: "tutorial_play_cover",
						Color:  1,
						Value:  "Tilt <I=G=MOVE> toward a wall: Cover (stick to wall)",
					},
					{
						LangId: "tutorial_play_dash",
						Color:  1,
						Value:  "<I=G=DASH>: Dash",
					},
					{
						LangId: "tutorial_play_evade",
						Color:  1,
						Value:  "<I=G=EVADE>: Quick dive",
					},
					{
						LangId: "",
						Color:  1,
						Value:  "Slowly tilt <I=G=MOVE> toward a cliff: Hang",
						Key:    1871380621,
					},
					{
						LangId: "",
						Color:  1,
						Value:  "While in cover, hold <I=G=STOCK>: Peek",
						Key:    2962995495,
					},
					{
						LangId: "tutorial_cancel",
						Color:  1,
						Value:  "<I=G=CANCEL>: Cancel",
					},
					{
						LangId: "tutorial_sonar",
						Color:  1,
						Value:  "<I=G=PAD_RS> while holding <I=G=CALL>: Use sonar (bio-detector)",
					},
					{
						LangId: "tutorial_order_buddy",
						Color:  1,
						Value:  "<I=G=PAD_RS> while holding <I=G=CALL>: Orders to buddy",
					},
					{
						LangId: "tutorial_order_child",
						Color:  1,
						Value:  "<I=G=PAD_RS> while holding <I=G=CALL>: Orders to children",
					},
					{
						LangId: "",
						Color:  1,
						Value:  "<I=G=STANCE>: In vehicle, hide/return upright",
						Key:    828898696,
					},
					{
						LangId: "",
						Color:  1,
						Value:  "<I=G=PAUSE>: View tips (pause screen)",
						Key:    1840317425,
					},
					{
						LangId: "tutorial_show_controller",
						Color:  1,
						Value:  "<I=G=PAUSE>: View controls (pause screen)",
					},
					{
						LangId: "tutorial_stance2",
						Color:  1,
						Value:  "<I=G=STANCE> Hold: Prone",
					},
					{
						LangId: "tutorial_stance3",
						Color:  1,
						Value:  "<I=G=STANCE>: Crouch",
					},
					{
						LangId: "tutorial_shield2",
						Color:  1,
						Value:  "<I=G=ACTION> while readying shield: Hide behind shield/exit cover",
					},
					{
						LangId: "tutorial_v_fps_tps",
						Color:  1,
						Value:  "<I=G=PAD_R1>: Switch camera (third-person/shoulder/first-person)",
					},
					{
						LangId: "tutorial_v_attack",
						Color:  1,
						Value:  "<I=G=PAD_L1>: Attack",
					},
				},
			},
			args: args{
				filename: "testdata/tpp_tutorial.eng.lng2.json",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Lng{}
			data, err := os.ReadFile(tt.args.filename)
			if err != nil {
				t.Fatal(err)
			}
			if err := l.UnmarshalJSON(data); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}

			if l.Header.Version != tt.fields.Header.Version {
				t.Errorf("Header.Version = %v, want %v", tt.fields.Header.Version, l.Header.Version)
			}

			if l.Header.Endianness != tt.fields.Header.Endianness {
				t.Errorf("Header.Endianness= %v, want %v", tt.fields.Header.Endianness, l.Header.Endianness)
			}

			if !slices.Equal(l.Entries, tt.fields.Entries) {
				t.Errorf("Entries = %+v, want %+v", l.Entries, tt.fields.Entries)
			}
		})
	}
}

func TestLng_Write(t *testing.T) {
	type fields struct {
		Header  Header
		Entries []Entry
		Keys    []Key
	}
	type args struct {
		filename     string
		filenameWant string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "eng",
			args: args{
				filename:     "testdata/tpp_tutorial.eng.lng2.json",
				filenameWant: "testdata/tpp_tutorial.eng.lng2",
			},
			wantErr: false,
		},
		{
			name: "jpn",
			args: args{
				filename:     "testdata/tpp_tutorial.jpn.lng2.json",
				filenameWant: "testdata/tpp_tutorial.jpn.lng2",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Lng{}
			src, err := os.ReadFile(tt.args.filename)
			if err != nil {
				t.Fatal(err)
			}
			if err = json.Unmarshal(src, &l); err != nil {
				t.Fatal(err)
			}

			buf := &util.ByteArrayReaderWriter{}
			if err = l.Write(buf); (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
			}

			want, err := os.ReadFile(tt.args.filenameWant)
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(buf.Bytes(), want) {
				t.Errorf("Write() got = %x, want %x", buf.Bytes(), want)
			}
		})
	}
}
