package crypto

import (
	"bytes"
	"testing"
)

func newTestCompressor(t *testing.T) *Compressor {
	t.Helper()
	c, err := NewCompressor()
	if err != nil {
		t.Fatalf("NewCompressor: %v", err)
	}
	t.Cleanup(c.Close)
	return c
}

func TestNewCompressor(t *testing.T) {
	c, err := NewCompressor()
	if err != nil {
		t.Fatalf("NewCompressor: %v", err)
	}
	c.Close()
}

func TestCompress_SmallPayload(t *testing.T) {
	c := newTestCompressor(t)

	small := []byte("short")
	result, compressed := c.Compress(small)

	if compressed {
		t.Error("small payload should not be compressed")
	}
	if !bytes.Equal(result, small) {
		t.Error("uncompressed result should equal input")
	}
}

func TestCompress_EmptyPayload(t *testing.T) {
	c := newTestCompressor(t)

	result, compressed := c.Compress(nil)
	if compressed {
		t.Error("nil payload should not be compressed")
	}
	if result != nil {
		t.Error("nil input should return nil")
	}

	result, compressed = c.Compress([]byte{})
	if compressed {
		t.Error("empty payload should not be compressed")
	}
	if len(result) != 0 {
		t.Error("empty input should return empty")
	}
}

func TestCompress_LargePayload(t *testing.T) {
	c := newTestCompressor(t)

	// Repetitive data compresses well
	large := bytes.Repeat([]byte("ABCDEFGHIJKLMNOP"), 100)
	result, compressed := c.Compress(large)

	if !compressed {
		t.Error("large repetitive payload should be compressed")
	}
	if len(result) >= len(large) {
		t.Errorf("compressed size %d should be smaller than original %d", len(result), len(large))
	}
}

func TestCompress_IncompressibleData(t *testing.T) {
	c := newTestCompressor(t)

	// Random-looking data doesn't compress well
	random := make([]byte, 256)
	for i := range random {
		random[i] = byte(i)
	}
	// XOR to make it less patterned
	for i := 1; i < len(random); i++ {
		random[i] ^= random[i-1] ^ random[(i*37)%256]
	}

	_, compressed := c.Compress(random)
	// May or may not compress depending on entropy — we just verify no crash
	_ = compressed
}

func TestCompressDecompress_RoundTrip(t *testing.T) {
	c := newTestCompressor(t)

	tests := []struct {
		name string
		data []byte
	}{
		{"repetitive text", bytes.Repeat([]byte("Hello, World! "), 100)},
		{"json-like", bytes.Repeat([]byte(`{"key":"value","num":42},`), 50)},
		{"binary pattern", bytes.Repeat([]byte{0x00, 0xFF, 0x80, 0x7F}, 200)},
		{"large zeros", make([]byte, 10000)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compressed, wasCompressed := c.Compress(tt.data)
			if !wasCompressed {
				t.Skip("data was not compressed (below threshold or no size reduction)")
			}

			decompressed, err := c.Decompress(compressed)
			if err != nil {
				t.Fatalf("Decompress: %v", err)
			}

			if !bytes.Equal(decompressed, tt.data) {
				t.Errorf("round-trip failed: got %d bytes, want %d bytes", len(decompressed), len(tt.data))
			}
		})
	}
}

func TestDecompress_InvalidData(t *testing.T) {
	c := newTestCompressor(t)

	_, err := c.Decompress([]byte("not valid zstd data"))
	if err == nil {
		t.Error("Decompress should fail on invalid data")
	}
}

func TestCompress_CompressionRatio(t *testing.T) {
	c := newTestCompressor(t)

	// Highly compressible: 16KB of zeros
	data := make([]byte, 16384)
	compressed, wasCompressed := c.Compress(data)

	if !wasCompressed {
		t.Fatal("zeros should compress")
	}

	ratio := float64(len(data)) / float64(len(compressed))
	if ratio < 10 {
		t.Errorf("compression ratio for zeros = %.1f, expected > 10", ratio)
	}
}

func TestCompressor_ConcurrentUse(t *testing.T) {
	c := newTestCompressor(t)
	data := bytes.Repeat([]byte("concurrent test data "), 100)

	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			compressed, ok := c.Compress(data)
			if !ok {
				return
			}
			decompressed, err := c.Decompress(compressed)
			if err != nil {
				t.Errorf("concurrent Decompress: %v", err)
				return
			}
			if !bytes.Equal(decompressed, data) {
				t.Error("concurrent round-trip failed")
			}
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
}

func BenchmarkCompress_1KB(b *testing.B) {
	c, _ := NewCompressor()
	defer c.Close()
	data := bytes.Repeat([]byte("benchmark data for compression "), 33)

	b.ResetTimer()
	for b.Loop() {
		c.Compress(data)
	}
}

func BenchmarkDecompress_1KB(b *testing.B) {
	c, _ := NewCompressor()
	defer c.Close()
	data := bytes.Repeat([]byte("benchmark data for compression "), 33)
	compressed, _ := c.Compress(data)

	b.ResetTimer()
	for b.Loop() {
		c.Decompress(compressed) //nolint:errcheck // benchmark
	}
}
