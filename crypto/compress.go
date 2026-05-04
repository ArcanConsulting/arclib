package crypto

import (
	"errors"
	"fmt"

	"github.com/klauspost/compress/zstd"
)

const (
	// MinCompressSize is the minimum payload size to consider compression.
	// Below this threshold, compression overhead exceeds any size savings.
	MinCompressSize = 64

	// ZSTDLevel is the default ZSTD compression level (1-19).
	// Level 3 balances speed and ratio for protocol payloads.
	ZSTDLevel = 3
)

var (
	// ErrDecompressFailed is returned when decompression fails.
	ErrDecompressFailed = errors.New("crypto: decompression failed")
)

// Compressor provides ZSTD compression and decompression for protocol payloads.
//
// Compressor reuses encoder/decoder instances for performance. It is safe
// for concurrent use — the underlying klauspost/zstd encoder and decoder
// are goroutine-safe.
type Compressor struct {
	enc *zstd.Encoder
	dec *zstd.Decoder
}

// NewCompressor creates a new ZSTD Compressor.
func NewCompressor() (*Compressor, error) {
	enc, err := zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.EncoderLevelFromZstd(ZSTDLevel)))
	if err != nil {
		return nil, fmt.Errorf("crypto: creating zstd encoder: %w", err)
	}

	dec, err := zstd.NewReader(nil)
	if err != nil {
		_ = enc.Close()
		return nil, fmt.Errorf("crypto: creating zstd decoder: %w", err)
	}

	return &Compressor{enc: enc, dec: dec}, nil
}

// Compress compresses data using ZSTD.
//
// Returns the original data unchanged if:
//   - The input is smaller than MinCompressSize (64 bytes)
//   - Compression does not reduce the size
//
// The second return value indicates whether compression was applied.
func (c *Compressor) Compress(data []byte) ([]byte, bool) {
	if len(data) < MinCompressSize {
		return data, false
	}

	compressed := c.enc.EncodeAll(data, nil)

	if len(compressed) >= len(data) {
		return data, false
	}

	return compressed, true
}

// Decompress decompresses ZSTD-compressed data.
func (c *Compressor) Decompress(data []byte) ([]byte, error) {
	result, err := c.dec.DecodeAll(data, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecompressFailed, err)
	}
	return result, nil
}

// Close releases the compressor resources.
func (c *Compressor) Close() {
	if c.enc != nil {
		_ = c.enc.Close()
	}
	if c.dec != nil {
		c.dec.Close()
	}
}
