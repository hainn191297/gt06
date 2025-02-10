package common

import "encoding/hex"

func ConvertToHexString(input []byte) string {
	if len(input) == 0 {
		return ""
	}
	return hex.EncodeToString(input)
}
