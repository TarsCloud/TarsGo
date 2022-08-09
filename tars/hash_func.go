package tars

// https://github.com/TarsCloud/TarsCpp/blob/master/util/include/util/tc_hash_fun.h

func HashString(str string) uint32 {
	var h uint32
	for _, c := range str {
		h = 5*h + uint32(c)
	}
	return h
}

func Hash(str string) uint32 {
	var h uint32
	for _, c := range str {
		h = (h << 4) + uint32(c)
		if g := h & 0xF0000000; g != 0 {
			h = h ^ (g >> 24)
			h = h ^ g
		}
	}
	return h
}

func HashNew(str string) uint32 {
	var value uint32
	for _, c := range str {
		value += uint32(c)
		value += value << 10
		value ^= value >> 6
	}
	value += value << 3
	value ^= value >> 11
	value += value << 15
	if value == 0 {
		return 1
	}
	return value
}

func MagicStringHash(str string) uint32 {
	return HashNew(str)
}
