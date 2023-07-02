package value

import (
	"fmt"

	"github.com/fxamacker/cbor/v2"
	"github.com/rafalb8/go-storage/encoding"
)

type cborCoder struct {
	cbor.EncMode
}

func CBORCoder() encoding.ValueCoder {
	opts := cbor.CanonicalEncOptions()
	opts.Time = cbor.TimeUnixDynamic

	enc, _ := opts.EncMode()
	return &cborCoder{
		EncMode: enc,
	}
}

// Option for CoderPair
func CBOR(c *encoding.CoderPair) {
	c.ValueCoder = CBORCoder()
}

func (c cborCoder) EncodeValue(val any) ([]byte, error) {
	return c.EncMode.Marshal(val)
}

func (c cborCoder) DecodeValue(data []byte, val any) error {
	if len(data) == 0 {
		return nil
	}

	if err := cbor.Valid(data); err != nil {
		return fmt.Errorf("invalid data: %w", err)
	}

	return cbor.Unmarshal(data, val)
}
