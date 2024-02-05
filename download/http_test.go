package download

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHTTPDownload(t *testing.T) {
	NewHttpDownloader()
	d := NewHttpDownloader().WithUrl("https://picsum.photos/id/237/200/300")
	err := d.Download("./test2.webp")
	require.NoError(t, err)
}
