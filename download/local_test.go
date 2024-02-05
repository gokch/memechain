package download

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLocalDownload(t *testing.T) {
	local := NewLocalDownloader()
	err := local.Download(context.Background(), "./test2.webp", "./test3.webp")
	require.NoError(t, err)
}
