package protocol

import "io"

// Adaptive chunking thresholds.
const (
	// SmallFileThreshold is the maximum file size that requires no chunking (1 MB).
	SmallFileThreshold = 1 << 20

	// MediumFileThreshold is the boundary between medium and large files (100 MB).
	MediumFileThreshold = 100 << 20

	// SmallChunkSize indicates no chunking for small files.
	SmallChunkSize = 0

	// MediumChunkSize is the chunk size for files between 1 MB and 100 MB.
	MediumChunkSize = 4 << 20

	// LargeChunkSize is the chunk size for files larger than 100 MB.
	LargeChunkSize = 16 << 20
)

// VFS OpCodes (0x0500-0x0515).
const (
	OpVFSMount   OpCode = 0x0500
	OpVFSUnmount OpCode = 0x0501
	OpVFSStat    OpCode = 0x0502
	OpVFSList    OpCode = 0x0503
	OpVFSRead    OpCode = 0x0504
	OpVFSWrite   OpCode = 0x0505
	OpVFSCreate  OpCode = 0x0506
	OpVFSDelete  OpCode = 0x0507
	OpVFSRename  OpCode = 0x0508
	OpVFSSync    OpCode = 0x0509

	OpChunkStore     OpCode = 0x0510
	OpChunkRetrieve  OpCode = 0x0511
	OpChunkVerify    OpCode = 0x0512
	OpChunkDelete    OpCode = 0x0513
	OpChunkLocate    OpCode = 0x0514
	OpChunkReplicate OpCode = 0x0515
)

func init() {
	RegisterOpCodes("vfs", map[OpCode]string{
		OpVFSMount:       "VFS_MOUNT",
		OpVFSUnmount:     "VFS_UNMOUNT",
		OpVFSStat:        "VFS_STAT",
		OpVFSList:        "VFS_LIST",
		OpVFSRead:        "VFS_READ",
		OpVFSWrite:       "VFS_WRITE",
		OpVFSCreate:      "VFS_CREATE",
		OpVFSDelete:      "VFS_DELETE",
		OpVFSRename:      "VFS_RENAME",
		OpVFSSync:        "VFS_SYNC",
		OpChunkStore:     "CHUNK_STORE",
		OpChunkRetrieve:  "CHUNK_RETRIEVE",
		OpChunkVerify:    "CHUNK_VERIFY",
		OpChunkDelete:    "CHUNK_DELETE",
		OpChunkLocate:    "CHUNK_LOCATE",
		OpChunkReplicate: "CHUNK_REPLICATE",
	})
}

// ChunkSize returns the appropriate chunk size for the given file size.
// Returns 0 for small files (no chunking needed).
func ChunkSize(fileSize int64) int {
	switch {
	case fileSize <= int64(SmallFileThreshold):
		return SmallChunkSize
	case fileSize <= int64(MediumFileThreshold):
		return MediumChunkSize
	default:
		return LargeChunkSize
	}
}

// ChunkCount returns the number of chunks for a given file size.
// Returns 1 for small files (entire file is a single chunk).
func ChunkCount(fileSize int64) int {
	if fileSize <= 0 {
		return 0
	}

	cs := ChunkSize(fileSize)
	if cs == 0 {
		return 1
	}

	count := int(fileSize) / cs
	if int(fileSize)%cs != 0 {
		count++
	}
	return count
}

// ChunkIterator iterates over chunks read from an io.Reader.
type ChunkIterator struct {
	reader       io.Reader
	chunkSize    int
	totalSize    int64
	currentChunk int
	totalChunks  int
}

// NewChunkIterator creates a new ChunkIterator for the given reader and file size.
// The chunk size is determined adaptively based on fileSize.
func NewChunkIterator(reader io.Reader, fileSize int64) *ChunkIterator {
	cs := ChunkSize(fileSize)
	if cs == 0 {
		// Small file: read entire file as one chunk.
		cs = int(fileSize)
	}

	return &ChunkIterator{
		reader:       reader,
		chunkSize:    cs,
		totalSize:    fileSize,
		currentChunk: 0,
		totalChunks:  ChunkCount(fileSize),
	}
}

// Next returns the next chunk of data, its index, and any error.
// Returns io.EOF when all chunks have been read.
func (ci *ChunkIterator) Next() (data []byte, index int, err error) {
	if ci.currentChunk >= ci.totalChunks {
		return nil, 0, io.EOF
	}

	// Determine how much to read for this chunk.
	readSize := ci.chunkSize
	remaining := int(ci.totalSize) - ci.currentChunk*ci.chunkSize
	if remaining < readSize {
		readSize = remaining
	}

	buf := make([]byte, readSize)
	n, err := io.ReadFull(ci.reader, buf)
	if err != nil && err != io.ErrUnexpectedEOF {
		return nil, 0, err
	}

	idx := ci.currentChunk
	ci.currentChunk++
	return buf[:n], idx, nil
}

// TotalChunks returns the total number of chunks.
func (ci *ChunkIterator) TotalChunks() int {
	return ci.totalChunks
}

// ChunkSizeValue returns the chunk size used by this iterator.
func (ci *ChunkIterator) ChunkSizeValue() int {
	return ci.chunkSize
}
