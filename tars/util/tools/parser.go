package tools

import (
	"strconv"
	"strings"
	"unicode"
)

const (
	// B byte
	B uint64 = 1
	// K kilobyte
	K uint64 = 1 << (10 * iota)
	// M megabyte
	M
	// G gigabyte
	G
	// T TeraByte
	T
	// P PetaByte
	P
	// E ExaByte
	E
)

var unitMap = map[string]uint64{
	"B":  B,
	"K":  K,
	"KB": K,
	"M":  M,
	"MB": M,
	"G":  G,
	"GB": G,
	"T":  T,
	"TB": T,
	"P":  P,
	"PB": P,
	"E":  E,
	"EB": E,
}

// ParseMegaByte translate xMB,xKB... to uint64 x (MB)
func ParseMegaByte(oriSize string) (ret uint64) {
	var defaultRotateSizeMB uint64 = 100
	if oriSize == "" {
		return defaultRotateSizeMB
	}
	sLogSize := ""
	sUnit := ""
	for idx, c := range oriSize {
		if !unicode.IsDigit(c) {
			sLogSize = strings.TrimSpace(oriSize[:idx])
			sUnit = strings.TrimSpace(oriSize[idx:])
			break
		}
	}
	if sLogSize == "" || sUnit == "" {
		return defaultRotateSizeMB
	}
	iLogSize, err := strconv.Atoi(sLogSize)
	if err != nil {
		return defaultRotateSizeMB
	}
	sUnit = strings.ToUpper(sUnit)
	iUnit, exists := unitMap[sUnit]
	if !exists {
		return defaultRotateSizeMB
	}
	ret = uint64(iLogSize) * iUnit / 1024 / 1024
	if ret == 0 {
		ret = defaultRotateSizeMB
	}
	return ret
}
