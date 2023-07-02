package key

import (
	"fmt"
	"strings"

	"github.com/rafalb8/go-storage/encoding"
)

type simple struct {
	encoding.Constants
}

func SimpleCoder() encoding.KeyCoder {
	return &simple{
		encoding.Constants{
			BucketKey:      [2]string{"[", "]"},
			Delimiter:      "//",
			TransactionKey: "[TX]",
		},
	}
}

func Simple(c *encoding.CoderPair) {
	c.KeyCoder = SimpleCoder()
}

func (c simple) Symbols() encoding.Constants {
	return c.Constants
}

func (c simple) EncodeKey(key ...string) string {
	if len(key) == 0 {
		return ""
	}
	return strings.Join(key, c.Delimiter)
}

func (c simple) DecodeKey(key string) []string {
	keys := strings.Split(key, c.Delimiter)
	return keys
}

func (c simple) EncodeBucket(key ...string) string {
	return fmt.Sprintf("%s%s%s", c.BucketKey[0], strings.Join(key, c.Delimiter), c.BucketKey[1])
}

func (c simple) DecodeBucket(key ...string) []string {
	keys := []string{}

	for _, k := range key {
		k = strings.Trim(k, c.BucketKey[0]+c.BucketKey[1])
		for _, b := range strings.Split(k, c.Delimiter) {
			keys = append(keys, strings.Trim(b, c.BucketKey[0]+c.BucketKey[1]))
		}
	}

	return keys
}
