package utils

import "strings"

// UpperFirstLetter Initial capitalization
func UpperFirstLetter(s string) string {
	if len(s) == 0 {
		return ""
	}
	if len(s) == 1 {
		return strings.ToUpper(string(s[0]))
	}
	return strings.ToUpper(string(s[0])) + s[1:]
}

func Path2ProtoName(path string) string {
	iBegin := strings.LastIndex(path, "/")
	if iBegin == -1 || iBegin >= len(path)-1 {
		iBegin = 0
	} else {
		iBegin++
	}
	iEnd := strings.LastIndex(path, ".tars")
	if iEnd == -1 {
		iEnd = len(path)
	}

	return path[iBegin:iEnd]
}
