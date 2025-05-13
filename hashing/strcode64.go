package hashing

import "github.com/unknown321/cityhash"

func StrCode64(s []byte) uint64 {
	sLen := len(s)
	if sLen < 1 {
		s = make([]byte, 1)
		sLen = 1
	}
	if s[sLen-1] != 0 {
		s = append(s, 0)
		sLen += 1
	}

	seed1 := uint(0)
	if sLen > 1 {
		seed1 = uint(s[0])<<16 + uint(sLen-1)
	}
	return cityhash.CityHash64WithSeeds(s, uint32(len(s)), StrCode32Seed0, uint64(seed1)) & 0xFFFFFFFFFFFF
}
