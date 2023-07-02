package key

import (
	"fmt"
	"strings"

	"github.com/rafalb8/go-storage/encoding"
)

type binary struct {
	encoding.Constants
}

func BinaryCoder() encoding.KeyCoder {
	return &binary{
		encoding.Constants{
			BucketKey:      [2]string{"\x1D", "\x1F"},
			Delimiter:      "\x1E",
			TransactionKey: "TX\x1C",
		},
	}
}

func Binary(c *encoding.CoderPair) {
	c.KeyCoder = BinaryCoder()
}

func (c binary) Symbols() encoding.Constants {
	return c.Constants
}

func (c binary) EncodeKey(key ...string) string {
	if len(key) == 0 {
		return ""
	}
	return strings.Join(key, c.Delimiter)
}

func (c binary) DecodeKey(key string) []string {
	keys := strings.Split(key, c.Delimiter)
	return keys
}

func (c binary) EncodeBucket(key ...string) string {
	return fmt.Sprintf("%s%s%s", c.BucketKey[0], strings.Join(key, c.Delimiter), c.BucketKey[1])
}

func (c binary) DecodeBucket(key ...string) []string {
	keys := []string{}

	for _, k := range key {
		k = strings.Trim(k, c.BucketKey[0]+c.BucketKey[1])
		for _, b := range strings.Split(k, c.Delimiter) {
			keys = append(keys, strings.Trim(b, c.BucketKey[0]+c.BucketKey[1]))
		}
	}

	return keys
}
