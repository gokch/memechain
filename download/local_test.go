package download

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLocalDownload(t *testing.T) {
	local := NewLocalDownloader().WithPath("./test2.webp")
	err := local.Download("./test3.webp")
	require.NoError(t, err)
}
