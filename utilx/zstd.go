package utilx

import (
	"sync"

	"github.com/klauspost/compress/zstd"
)

const (
	DefaultZSTDCompressionLevel = 3
)

var (
	decoder *zstd.Decoder
	encoder *zstd.Encoder

	encOnce, decOnce sync.Once
)

// ZSTDDecompress decompresses a block using ZSTD algorithm.
func ZSTDDecompress(dst, src []byte) ([]byte, error) {
	var err error
	decOnce.Do(func() {
		decoder, err = zstd.NewReader(nil)
	})
	if err != nil {
		return nil, err
	}
	return decoder.DecodeAll(src, dst[:0])
}

// ZSTDCompress compresses a block using ZSTD algorithm.
func ZSTDCompress(dst, src []byte, compressionLevel int) ([]byte, error) {
	if dst == nil {
		dst = make([]byte, ZSTDCompressBound(len(src)))
	}
	var err error
	encOnce.Do(func() {
		level := zstd.EncoderLevelFromZstd(compressionLevel)
		encoder, err = zstd.NewWriter(nil, zstd.WithEncoderLevel(level))
	})
	if err != nil {
		return nil, err
	}
	return encoder.EncodeAll(src, dst[:0]), nil
}

// ZSTDCompressBound returns the worst case size needed for a destination buffer.
// Klauspost ZSTD library does not provide any API for Compression Bound. This
// calculation is based on the DataDog ZSTD library.
// See https://pkg.go.dev/github.com/DataDog/zstd#CompressBound
func ZSTDCompressBound(srcSize int) int {
	lowLimit := 128 << 10 // 128 kB
	var margin int
	if srcSize < lowLimit {
		margin = (lowLimit - srcSize) >> 11
	}
	return srcSize + (srcSize >> 8) + margin
}
