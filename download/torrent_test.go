package download

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTorrentDownload(t *testing.T) {
	d, err := NewTorrentDownloader()
	require.NoError(t, err)
	err = d.Download(context.Background(), "magnet:?xt=urn:btih:KRWPCX3SJUM4IMM4YF5RPHL6ANPYTQPU", "./")
	require.NoError(t, err)
}
