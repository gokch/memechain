package download

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTorrentDownload(t *testing.T) {
	d := NewTorrentDownloader().WithURI("magnet:?xt=urn:btih:KRWPCX3SJUM4IMM4YF5RPHL6ANPYTQPU")
	err := d.Download("")
	require.NoError(t, err)

}
