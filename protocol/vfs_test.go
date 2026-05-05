package protocol

import (
	"bytes"
	"io"
	"testing"
)

func TestChunkSize(t *testing.T) {
	tests := []struct {
		name     string
		fileSize int64
		want     int
	}{
		{"zero", 0, SmallChunkSize},
		{"500KB", 500 << 10, SmallChunkSize},
		{"exactly 1MB", SmallFileThreshold, SmallChunkSize},
		{"50MB", 50 << 20, MediumChunkSize},
		{"exactly 100MB", MediumFileThreshold, MediumChunkSize},
		{"500MB", 500 << 20, LargeChunkSize},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ChunkSize(tt.fileSize)
			if got != tt.want {
				t.Errorf("ChunkSize(%d) = %d, want %d", tt.fileSize, got, tt.want)
			}
		})
	}
}

func TestChunkCount(t *testing.T) {
	tests := []struct {
		name     string
		fileSize int64
		want     int
	}{
		{"zero", 0, 0},
		{"500KB", 500 << 10, 1},
		{"exactly 1MB", SmallFileThreshold, 1},
		{"50MB", 50 << 20, 13},                  // 50/4 = 12.5 -> 13
		{"exactly 100MB", 100 << 20, 25},        // 100/4 = 25
		{"500MB", 500 << 20, 32},                // 500/16 = 31.25 -> 32
		{"exactly 16MB multiple", 64 << 20, 16}, // 64/4 = 16
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ChunkCount(tt.fileSize)
			if got != tt.want {
				t.Errorf("ChunkCount(%d) = %d, want %d", tt.fileSize, got, tt.want)
			}
		})
	}
}

func TestChunkIteratorSmallFile(t *testing.T) {
	// Small file: entire content returned as a single chunk.
	data := make([]byte, 512<<10) // 512 KB
	for i := range data {
		data[i] = byte(i % 256)
	}

	reader := bytes.NewReader(data)
	iter := NewChunkIterator(reader, int64(len(data)))

	if iter.TotalChunks() != 1 {
		t.Fatalf("TotalChunks() = %d, want 1", iter.TotalChunks())
	}

	chunk, idx, err := iter.Next()
	if err != nil {
		t.Fatalf("Next() error: %v", err)
	}
	if idx != 0 {
		t.Errorf("chunk index = %d, want 0", idx)
	}
	if !bytes.Equal(chunk, data) {
		t.Errorf("chunk data does not match original")
	}

	// Next call should return EOF.
	_, _, err = iter.Next()
	if err != io.EOF {
		t.Errorf("expected io.EOF after last chunk, got %v", err)
	}
}

func TestChunkIteratorMediumFile(t *testing.T) {
	// Medium file: 10 MB -> should produce 4 MB chunks.
	fileSize := int64(10 << 20)
	data := make([]byte, fileSize)
	for i := range data {
		data[i] = byte(i % 251) // prime modulus for non-trivial pattern
	}

	reader := bytes.NewReader(data)
	iter := NewChunkIterator(reader, fileSize)

	expectedChunks := 3 // 10/4 = 2.5 -> 3
	if iter.TotalChunks() != expectedChunks {
		t.Fatalf("TotalChunks() = %d, want %d", iter.TotalChunks(), expectedChunks)
	}

	if iter.ChunkSizeValue() != MediumChunkSize {
		t.Errorf("ChunkSizeValue() = %d, want %d", iter.ChunkSizeValue(), MediumChunkSize)
	}

	reassembled := make([]byte, 0, fileSize)
	for i := range expectedChunks {
		chunk, idx, err := iter.Next()
		if err != nil {
			t.Fatalf("Next() chunk %d error: %v", i, err)
		}
		if idx != i {
			t.Errorf("chunk index = %d, want %d", idx, i)
		}
		reassembled = append(reassembled, chunk...)
	}

	if !bytes.Equal(reassembled, data) {
		t.Errorf("reassembled data does not match original")
	}

	// Should be EOF now.
	_, _, err := iter.Next()
	if err != io.EOF {
		t.Errorf("expected io.EOF, got %v", err)
	}
}

func TestChunkIteratorExactBoundary(t *testing.T) {
	// File size is an exact multiple of the chunk size: 8 MB = 2 * 4 MB.
	fileSize := int64(8 << 20)
	data := make([]byte, fileSize)
	for i := range data {
		data[i] = byte(i % 199)
	}

	reader := bytes.NewReader(data)
	iter := NewChunkIterator(reader, fileSize)

	expectedChunks := 2 // 8/4 = 2 exactly
	if iter.TotalChunks() != expectedChunks {
		t.Fatalf("TotalChunks() = %d, want %d", iter.TotalChunks(), expectedChunks)
	}

	reassembled := make([]byte, 0, fileSize)
	for i := range expectedChunks {
		chunk, idx, err := iter.Next()
		if err != nil {
			t.Fatalf("Next() chunk %d error: %v", i, err)
		}
		if idx != i {
			t.Errorf("chunk index = %d, want %d", idx, i)
		}
		if len(chunk) != MediumChunkSize {
			t.Errorf("chunk %d size = %d, want %d", i, len(chunk), MediumChunkSize)
		}
		reassembled = append(reassembled, chunk...)
	}

	if !bytes.Equal(reassembled, data) {
		t.Errorf("reassembled data does not match original")
	}
}

func TestVFSOpCodesRegistered(t *testing.T) {
	tests := []struct {
		op   OpCode
		name string
	}{
		{OpVFSMount, "VFS_MOUNT"},
		{OpVFSUnmount, "VFS_UNMOUNT"},
		{OpVFSStat, "VFS_STAT"},
		{OpVFSList, "VFS_LIST"},
		{OpVFSRead, "VFS_READ"},
		{OpVFSWrite, "VFS_WRITE"},
		{OpVFSCreate, "VFS_CREATE"},
		{OpVFSDelete, "VFS_DELETE"},
		{OpVFSRename, "VFS_RENAME"},
		{OpVFSSync, "VFS_SYNC"},
		{OpChunkStore, "CHUNK_STORE"},
		{OpChunkRetrieve, "CHUNK_RETRIEVE"},
		{OpChunkVerify, "CHUNK_VERIFY"},
		{OpChunkDelete, "CHUNK_DELETE"},
		{OpChunkLocate, "CHUNK_LOCATE"},
		{OpChunkReplicate, "CHUNK_REPLICATE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LookupOpCode(tt.op)
			if got != tt.name {
				t.Errorf("LookupOpCode(0x%04X) = %q, want %q", uint16(tt.op), got, tt.name)
			}
			// Also verify String() works.
			if s := tt.op.String(); s != tt.name {
				t.Errorf("OpCode(0x%04X).String() = %q, want %q", uint16(tt.op), s, tt.name)
			}
		})
	}
}
