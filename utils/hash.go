package utils

import (
	"crypto/md5"
	"encoding/hex"
)

// HashMD5 takes a string and outputs the md5 hash of it
func HashMD5(str string) string {
	hasher := md5.New()
	hasher.Write([]byte(str))

	return hex.EncodeToString(hasher.Sum(nil))
}
