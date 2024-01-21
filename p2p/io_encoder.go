package p2p

import (
	"bytes"
	"io"
	"github.com/gokch/memechain/utilx"
	"strings"

	"github.com/klauspost/compress/zstd"
)

var EncodingPriorMap = map[string]int{
	EncodingTypeNone: 9,
	EncodingTypeZstd: 1,
}

func GetPreferredEncodingType(acceptEncodingHeader []string) string {
	encodingType := EncodingTypeNone

	for _, v := range acceptEncodingHeader {
		for _, vv := range strings.Split(v, ",") {
			vv = strings.TrimSpace(vv)
			priority, ok := EncodingPriorMap[vv]
			if ok && priority < EncodingPriorMap[encodingType] {
				encodingType = vv
			}
		}
	}

	return encodingType
}

func IsAvailableEncoding(encodingType *string) bool {
	if encodingType == nil {
		return false
	}

	if *encodingType == "" {
		*encodingType = EncodingTypeNone
		return true
	}

	_, ok := EncodingPriorMap[*encodingType]
	return ok
}

type EncodingReadCloser struct {
	reader    io.ReadCloser
	closeFunc func() error
}

func (e *EncodingReadCloser) Close() error {
	return e.closeFunc()
}

func (e *EncodingReadCloser) Read(p []byte) (int, error) {
	return e.reader.Read(p)
}

func NewEncodingReadCloser(reader io.ReadCloser, encodingType string) io.ReadCloser {
	switch encodingType {
	case EncodingTypeZstd:
		zstdReadCloser, err := zstd.NewReader(reader)
		if err != nil {
			reader.Close()
		}
		ioReadCloser := zstdReadCloser.IOReadCloser()
		return &EncodingReadCloser{
			reader: ioReadCloser,
			closeFunc: func() error {
				err := ioReadCloser.Close()
				if err != nil {
					reader.Close()
					return err
				}

				return reader.Close()
			},
		}
	default:
		return &EncodingReadCloser{
			reader: reader,
			closeFunc: func() error {
				return reader.Close()
			},
		}
	}
}

func NewBufferEncodingReader(data []byte, encodingType string) (io.Reader, int64, error) {
	switch encodingType {
	case EncodingTypeZstd:
		var err error
		data, err = utilx.ZSTDCompress(nil, data, utilx.DefaultZSTDCompressionLevel)
		if err != nil {
			return nil, 0, err
		}
	}
	return bytes.NewBuffer(data), int64(len(data)), nil
}

func NewEncodingWriteCloser(writer io.Writer, encodingType string) io.WriteCloser {
	switch encodingType {
	case EncodingTypeZstd:
		enc, err := zstd.NewWriter(writer, nil)
		if err != nil {
			return nil
		}
		return enc
	}

	return WNopCloser(writer)
}

type wnopCloser struct {
	io.Writer
}

func (wnopCloser) Close() error { return nil }

func WNopCloser(r io.Writer) io.WriteCloser {
	return wnopCloser{r}
}

type limitReadCloser struct {
	reader    io.Reader
	closeFunc func() error
}

func (l *limitReadCloser) Close() error {
	return l.closeFunc()
}

func (l *limitReadCloser) Read(p []byte) (int, error) {
	return l.reader.Read(p)
}

func LimitReadCloser(reader io.ReadCloser, n int64) io.ReadCloser {
	return &limitReadCloser{
		reader: io.LimitReader(reader, n),
		closeFunc: func() error {
			return reader.Close()
		},
	}
}
