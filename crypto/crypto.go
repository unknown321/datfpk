package crypto

const Magic1 = 0xA0F8EFE6
const Magic2 = 0xE3F8EFE6

func GetHeaderSize(encryption uint32) int {
	headerSize := 0
	switch encryption {
	case Magic1:
		headerSize = 8
	case Magic2:
		headerSize = 16
	}

	return headerSize
}
