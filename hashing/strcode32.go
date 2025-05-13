package hashing

import (
	"encoding/binary"
	"github.com/unknown321/cityhash"
	"slices"
	"unsafe"
)

const l = unsafe.Sizeof(uint64(0))
const StrCode32Seed0 = uint64(0x9ae16a3b2f90404f)

func StrCode32(s []byte) uint64 {
	s1 := make([]byte, len(s))
	copy(s1, s)
	slices.Reverse(s1)
	for len(s1) < int(l) {
		s1 = append(s1, 0)
	}

	sd1 := make([]byte, l)
	copy(sd1, s1[0:l])

	seed1 := binary.LittleEndian.Uint64(sd1)
	return cityhash.CityHash64WithSeeds(s, uint32(len(s)), StrCode32Seed0, seed1) & 0x3FFFFFFFFFFFF
}
