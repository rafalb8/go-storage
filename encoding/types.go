package encoding

type Constants struct {
	Delimiter      string
	TransactionKey string
}

type CoderOpts func(*CoderPair)

// Encoder / Decoder interface
type Coder interface {
	KeyCoder
	ValueCoder
}

type KeyCoder interface {
	// Returns struct with all coder constants
	Symbols() Constants

	EncodeKey(key ...string) string
	DecodeKey(key string) []string

	// Encode/Decode bucket key

	EncodeBucket(key ...string) string
	DecodeBucket(bucket ...string) []string
}

type ValueCoder interface {
	EncodeValue(val any) ([]byte, error)
	DecodeValue(data []byte, val any) error
}

// Interface for custom EncodeValue implementation
type CustomValueEncoder interface {
	EncodeValue(coder ValueCoder) ([]byte, error)
}

// Interface for custom DecodeValue implementation
type CustomValueDecoder interface {
	DecodeValue(coder ValueCoder, data []byte) error
}
