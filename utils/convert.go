package utils

import (
	"bytes"
	"encoding/gob"
)

// ToBytes converts an interface of arbitrary type to byte array
func ToBytes(key interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
