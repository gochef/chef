package utils

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"strings"
	"unicode"
)

var (
	strChars = []byte("ABCDEFGHIJKLMNOPQRSTUVXYZabcdefghijklmnopqrstuvwxyz1234567890")
	numChars = []byte("1234567890")
)

// RandomBytes returns a new random bytes of the provided length
func RandomBytes(length int, chars []byte) ([]byte, error) {
	if length == 0 {
		return []byte(""), nil
	}

	clen := len(chars)
	maxrb := 255 - (256 % clen)

	b := make([]byte, length)
	r := make([]byte, length+(length/4))
	i := 0
	for {
		if _, err := rand.Read(r); err != nil {
			return nil, fmt.Errorf("chef: error reading random bytes: %s", err.Error())
		}

		for _, rb := range r {
			c := int(rb)
			if c > maxrb {
				// Skip this number to avoid modulo bias.
				continue
			}

			b[i] = chars[c%clen]
			i++
			if i == length {
				return b, nil
			}
		}
	}
}

// RandomString returns a new random string of the provided length
func RandomString(length int) (string, error) {
	b, err := RandomBytes(length, strChars)
	return string(b), err
}

func RandomNumberString(length int) (string, error) {
	b, err := RandomBytes(length, numChars)
	return string(b), err
}

// IsUpperCase checks if a character is upper case. More precisely it evaluates if it is
// in the range of ASCII characters 'A' to 'Z'.
func IsUpperCase(ch rune) bool {
	return ch >= 'A' && ch <= 'Z'
}

// ToUpperCase converts a character in the range of ASCII characters 'a' to 'z' to its upper
// case counterpart. Other characters remain the same.
func ToUpperCase(ch rune) rune {
	if ch >= 'a' && ch <= 'z' {
		return ch - 32
	}
	return ch
}

// ToLowerCase converts a character in the range of ASCII characters 'A' to 'Z' to its lower
// case counterpart. Other characters remain the same.
func ToLowerCase(ch rune) rune {
	if ch >= 'A' && ch <= 'Z' {
		return ch + 32
	}
	return ch
}

// isSpace checks if a character is some kind of whitespace.
func isSpace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

// isDelimiter checks if a character is some kind of whitespace or '_' or '-'.
func isDelimiter(ch rune) bool {
	return ch == '-' || ch == '_' || isSpace(ch)
}

// StrToCamelCase converts a string to its camel case representation
func StrToCamelCase(str string) string {
	str = strings.TrimSpace(str)
	buffer := make([]rune, 0, len(str))

	var prev rune
	for _, curr := range str {
		if !isDelimiter(curr) {
			if isDelimiter(prev) || (prev == 0) {
				buffer = append(buffer, ToUpperCase(curr))
			} else {
				buffer = append(buffer, ToLowerCase(curr))
			}
		}
		prev = curr
	}

	return string(buffer)
}

// StrToSnakeCase converts a string to its snake case representation
func StrToSnakeCase(str string) string {
	buf := new(bytes.Buffer)

	runes := []rune(str)
	for i := 0; i < len(runes); i++ {
		buf.WriteRune(unicode.ToLower(runes[i]))
		if i != len(runes)-1 && unicode.IsUpper(runes[i+1]) &&
			(unicode.IsLower(runes[i]) || unicode.IsDigit(runes[i]) ||
				(i != len(runes)-2 && unicode.IsLower(runes[i+2]))) {
			buf.WriteRune('_')
		}
	}

	return buf.String()
}

// StrStartsWith checks if a string starts with a substring
func StrStartsWith(str, prefix string) bool {
	return strings.HasPrefix(str, prefix)
}

// StrEndsWith checks if a string ends with a substring
func StrEndsWith(str, suffix string) bool {
	return strings.HasSuffix(str, suffix)
}

// StrContains checks if a string contains a substring
func StrContains(str, substr string) bool {
	return strings.Contains(str, substr)
}

type converter func(string) string

// Convert converts a list of string using the passed converter function
func Convert(s []string, c converter) []string {
	out := []string{}
	for _, i := range s {
		out = append(out, c(i))
	}
	return out

}

// DelimString adds the passed delimeter after every n characters in the passed string
func DelimitString(str, delim string, chunkLength int) string {
	var chunks []string
	runes := []rune(str)

	if len(runes) == 0 {
		return str
	}

	for i := 0; i < len(runes); i += chunkLength {
		nn := i + chunkLength
		if nn > len(runes) {
			nn = len(runes)
		}

		chunks = append(chunks, string(runes[i:nn]))
	}

	return strings.Join(chunks, delim)
}
