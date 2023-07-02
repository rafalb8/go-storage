package value

import (
	"encoding/json"
	"fmt"

	"github.com/rafalb8/go-storage/encoding"
)

type jsonCoder struct {
}

func JSONCoder() encoding.ValueCoder {
	return &jsonCoder{}
}

// Option for CoderPair
func JSON(c *encoding.CoderPair) {
	c.ValueCoder = JSONCoder()
}

func (c jsonCoder) EncodeValue(val any) ([]byte, error) {
	return json.Marshal(val)
}

func (c jsonCoder) DecodeValue(data []byte, val any) error {
	if len(data) == 0 {
		return nil
	}

	err := json.Unmarshal(data, val)
	if err != nil {
		if !json.Valid(data) {
			return fmt.Errorf("invalid data: %w", err)
		}
	}
	return err
}
