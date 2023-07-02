package value

import (
	"fmt"

	"github.com/fxamacker/cbor/v2"
	"github.com/rafalb8/go-storage/encoding"
	"github.com/rafalb8/go-storage/internal"
)

var (
	log = internal.Logger()
)

type cborCoder struct {
	cbor.EncMode
}

func CBORCoder() encoding.ValueCoder {
	opts := cbor.CanonicalEncOptions()
	opts.Time = cbor.TimeUnixDynamic

	enc, err := opts.EncMode()
	if err != nil {
		log.Fatal(err)
	}

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
