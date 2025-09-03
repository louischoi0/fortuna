package util

import (
	"fortuna/crypto"
	"strings"
)

const CONCAT_HASH_DELIMTER = ":"

func ConcatHash(strs ...string) string {
	var cursor strings.Builder	

	for _, str := range(strs) {
		cursor.WriteString(str)
		cursor.WriteString(CONCAT_HASH_DELIMTER)
	}

	return crypto.SHA256(cursor.String())
}
