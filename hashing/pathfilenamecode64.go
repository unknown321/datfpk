package hashing

import (
	"bytes"
	"github.com/unknown321/cityhash"
	"strings"
)

//go:generate go run generate/generate.go
//go:generate go fmt extension.go

const MetaFlag = 0x4000000000000

func HashFileNameWithExtension(s string) uint64 {
	ext := ""
	dotIndex := strings.LastIndex(s, ".")
	if dotIndex > 0 {
		ext = s[dotIndex+1:]
	}

	eh, ok := Extensions[ext]
	if !ok {
		eh = 0
	}

	f := HashFileName(s, true)
	out := (eh << 51) | f
	return out
}

func HashFileName(s string, removeExtension bool) uint64 {
	trimmed := strings.TrimPrefix(s, "/Assets/")
	trimmed = strings.TrimPrefix(trimmed, "/")

	isMeta := false
	if !strings.HasPrefix(s, "/Assets/") || strings.HasPrefix(trimmed, "tpptest") {
		isMeta = true
	}

	if removeExtension {
		dotIndex := strings.LastIndex(trimmed, ".")
		if dotIndex > 0 {
			trimmed = trimmed[:dotIndex]
		}
	}

	tt := StrCode32([]byte(trimmed))
	if isMeta {
		tt = tt | MetaFlag
	}

	return tt
}

func HashFileNameLegacy(s []byte, removeExtension bool) uint64 {
	if removeExtension {
		dotIndex := bytes.LastIndex(s, []byte("."))
		if dotIndex > 0 {
			s = s[:dotIndex]
		}
	}

	seed1 := 0
	if len(s) > 0 {
		seed1 = int(s[0])<<16 + len(s)
	}

	s = append(s, byte(0))

	res := cityhash.CityHash64WithSeeds(s, uint32(len(s)), StrCode32Seed0, uint64(seed1)) & 0xFFFFFFFFFFFF

	return res
}

func PathHashFromHash(hash uint64) uint64 {
	return hash & 0x3FFFFFFFFFFFF
}

func ExtHashFromHash(hash uint64) uint64 {
	return hash >> 51
}

func JustAddExtension(hash uint64, ext string) uint64 {
	eh, ok := Extensions[ext]
	if !ok {
		eh = 0
	}
	return (eh << 51) | hash
}
