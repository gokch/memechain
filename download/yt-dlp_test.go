package download

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestYtDlpDownload(t *testing.T) {
	d := NewYtDlpDownloader().WithUrl("https://www.youtube.com/watch?v=wILX71HE3e0")
	err := d.Download("./youtube.mp4")
	require.NoError(t, err)

	err = d.Download("./")
	require.NoError(t, err)
}
