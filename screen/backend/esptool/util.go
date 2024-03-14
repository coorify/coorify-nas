package esptool

import "fmt"

func uint32ToBytes(value uint32) []byte {
	return []byte{byte(value & 0xFF),
		byte((value >> 8) & 0xFF),
		byte((value >> 16) & 0xFF),
		byte((value >> 24) & 0xFF),
	}
}

func uint16ToBytes(value uint16) []byte {
	return []byte{byte(value & 0xFF), byte((value >> 8) & 0xFF)}
}

func bytesToUint32(value []byte) uint32 {
	return uint32(value[0]) | (uint32(value[1]) << 8) | (uint32(value[2]) << 16) | (uint32(value[3]) << 24)
}

func checksum(data []byte) uint32 {
	state := uint32(0xEF)

	for _, d := range data {
		state ^= uint32(d)
	}

	return state
}

func hexConvert(data []byte) string {
	dlen := len(data)

	if dlen > 16 {
		result := "\n"
		for dlen > 0 {
			slen := 16
			if slen > dlen {
				slen = dlen
			}

			result += fmt.Sprintf("%s\n", hexify(data[0:slen]))

			data = data[slen:]
			dlen = len(data)
		}

		return result
	} else {
		return hexify(data)
	}
}

func hexify(data []byte) string {
	result := ""
	for _, v := range data {
		result += fmt.Sprintf("%02X", v)
	}
	return result
}
