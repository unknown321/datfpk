package fox2

import (
	"bytes"
	"datfpk/fox2/containers"
	"datfpk/fox2/datatypes/fox"
	"datfpk/hashing"
	"encoding/xml"
	"fmt"
	"reflect"
	"testing"
)

func TestProperty_UnmarshalXML(t *testing.T) {
	type want struct {
		Header    PropertyHeader
		Value     IFoxContainer
		NameValue string
	}
	type args struct {
		in string
	}
	tests := []struct {
		name string
		want want
		args args
	}{
		{
			name: "",
			want: want{
				Header: PropertyHeader{
					NameHash:      0,
					DataType:      fox.FString,
					ContainerType: containers.StaticArray,
					ValueCount:    2,
					Offset:        0,
					Size:          0,
				},
				Value: &containers.FoxStaticArray{
					Data: []fox.DataType{
						&fox.String{
							Hash:  0,
							Value: "123",
						},
						&fox.String{
							Hash:  0,
							Value: "uhh",
						},
					},
				},
				NameValue: "name",
			},
			args: args{
				in: `<property name="name" type="String" container="StaticArray" arraySize="2">
  <containerEntry>123</containerEntry>
  <containerEntry>uhh</containerEntry>
</property>`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Property{
				Header:    tt.want.Header,
				Value:     tt.want.Value,
				NameValue: tt.want.NameValue,
			}

			if err := xml.Unmarshal([]byte(tt.args.in), p); err != nil {
				t.Fatalf("%s", err.Error())
			}

			if !reflect.DeepEqual(tt.want.Header, p.Header) {
				fmt.Printf("have: %+v\nwant %+v", p.Header, tt.want.Header)
				t.Fatalf("not equal")
			}

			// don't want to bother with type checking
			um, err := xml.MarshalIndent(p, "", "  ")
			if err != nil {
				t.Fatalf("marshal fail: %s", err.Error())
			}
			if string(um) != tt.args.in {
				t.Fatalf("not equal: %s", um)
			}
		})
	}
}

func TestProperty_MarshalXML(t *testing.T) {
	//if err := Init("../fox_dictionary.txt"); err != nil {
	//	t.Fatalf("%s", err.Error())
	//}

	fox2dict[hashing.StrCode64([]byte("name"))] = "name"

	p := Property{
		Header: PropertyHeader{
			NameHash:      0x3391ed17a03a,
			DataType:      11,
			ContainerType: containers.StaticArray,
			ValueCount:    2,
		},
		Value: nil,
	}

	p.NameValue = fox2dict[p.Header.NameHash]

	cc, err := CreateTypedContainer(fox.FString, containers.StaticArray, 2)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	p.Value = cc
	d := (p.Value).(*containers.FoxStaticArray)
	d.Data[0] = &fox.String{Value: "123", Hash: 0x123456}
	d.Data[1] = &fox.String{
		Hash:  0,
		Value: "uhh",
	}

	b, err := xml.MarshalIndent(&p, "", "  ")
	if err != nil {
		t.Fatalf("%s", err.Error())
	}

	want := []byte(`<property name="name" type="String" container="StaticArray" arraySize="2">
  <containerEntry>123</containerEntry>
  <containerEntry>uhh</containerEntry>
</property>`)
	if bytes.Compare(b, want) != 0 {
		t.Fatalf("have \n%s\n want \n%s\n", b, want)
	}
}
