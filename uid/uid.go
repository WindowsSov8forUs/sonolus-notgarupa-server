package uid

import (
	"crypto/rand"
	"fmt"
	"io"
	"strings"
	"time"
)

const (
	Alphabet     = "23456789abcdefghijkmnpqrstuvwxyz"
	timeLength   = 9
	randomLength = 6
)

var enginePrefixes = map[string]string{
	"notgarupa":          "notgarupa-",
	"notgarupa-habahiro": "habahiro-",
}

func Generate(engine string, now time.Time) (string, error) {
	prefix, ok := enginePrefixes[engine]
	if !ok {
		return "", fmt.Errorf("missing level uid prefix for engine %q", engine)
	}
	random, err := randomBase32(randomLength)
	if err != nil {
		return "", err
	}
	return prefix + encodeBase32(now.UnixMilli(), timeLength) + "_" + random, nil
}

func encodeBase32(value int64, length int) string {
	if value < 0 {
		value = 0
	}
	if value == 0 {
		return strings.Repeat(string(Alphabet[0]), length)
	}
	var reversed []byte
	base := int64(len(Alphabet))
	for value > 0 {
		reversed = append(reversed, Alphabet[value%base])
		value /= base
	}

	encoded := make([]byte, max(length, len(reversed)))
	for i := range encoded {
		encoded[i] = Alphabet[0]
	}
	for i, b := range reversed {
		encoded[len(encoded)-1-i] = b
	}
	return string(encoded)
}

func randomBase32(length int) (string, error) {
	buf := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, buf); err != nil {
		return "", err
	}
	for i, value := range buf {
		buf[i] = Alphabet[value&31]
	}
	return string(buf), nil
}
