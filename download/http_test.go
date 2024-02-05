package download

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHTTPDownload(t *testing.T) {
	d, err := NewHttpDownloader()
	require.NoError(t, err)
	err = d.Download(context.Background(), "https://picsum.photos/id/237/200/300", "./test2.webp")
	require.NoError(t, err)
}
