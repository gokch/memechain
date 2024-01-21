package types

import (
	"io/fs"
	"sync/atomic"

	"github.com/gokch/memechain/utilx"
)

type Dir struct {
	Name string      `json:"name"`
	Perm fs.FileMode `json:"perm"`
}

type Symlink struct {
	Name      string `json:"name"`
	SymlinkTo string `json:"symlinkTo"`
}

type EmptyFile struct {
	Name string      `json:"name"`
	Perm fs.FileMode `json:"perm"`
}

type File struct {
	Name string      `json:"name"`
	Size int64       `json:"size"`
	Perm fs.FileMode `json:"perm"`
}

type AvailableChunkList struct {
	Timestamp              string `json:"timestamp"`
	IsCompleted            bool   `json:"isCompleted"`
	TransferChunkIndexList []int  `json:"transferChunkIndexList"`
}

type FileChunk struct {
	FileIndex  int   `json:"index"`
	FromOffset int64 `json:"from"`
	ToOffset   int64 `json:"to"`
}

const (
	ChunkStatusPending = iota
	ChunkStatusDownloading
	ChunkStatusDone
)

type TransferChunk struct {
	FileChunks []FileChunk `json:"fileChunks"`
	// chunk 의 ChunkStatus (ChunkStatusPending / ChunkStatusDownloading / ChunkStatusDone)
	Status        atomic.Int32     `json:"-"`
	LastUpdatedAt utilx.AtomicTime `json:"-"`
}

func (t *TransferChunk) Size() int64 {
	var length int64
	for _, fileChunk := range t.FileChunks {
		length += (fileChunk.ToOffset - fileChunk.FromOffset)
	}
	return length
}
