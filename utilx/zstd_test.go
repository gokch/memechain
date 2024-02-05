package utilx

import (
	"testing"

	"github.com/funkygao/assert"
	"github.com/stretchr/testify/require"
)

func TestZSTDCompressDecompress(t *testing.T) {
	for _, test := range []struct {
		name             string
		compressionLevel int
		origin           []byte
	}{
		{"empty", 1, []byte{}},
		{"short", 2, []byte("hello world")},
		{"long", 3, []byte("hello world, hello world, hello world, hello world, hello world, hello world, hello world, hello world, hello world, hello world")},
	} {
		dst := make([]byte, 0)
		compress, err := ZSTDCompress(dst, test.origin, test.compressionLevel)
		require.NoErrorf(t, err, "test %s", test.name)
		recover, err := ZSTDDecompress(dst, compress)
		require.NoErrorf(t, err, "test %s", test.name)
		assert.Equalf(t, test.origin, recover, "test %s", test.name)
	}
}
