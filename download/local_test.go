package download

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLocalDownload(t *testing.T) {
	local, err := NewLocalDownloader()
	require.NoError(t, err)
	err = local.Download(context.Background(), "./test2.webp", "./test3.webp")
	require.NoError(t, err)
}
