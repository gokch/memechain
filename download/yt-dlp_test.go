package download

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestYtDlpDownload(t *testing.T) {
	d := NewYtDlpDownloader()
	err := d.Download(context.Background(), "https://www.youtube.com/watch?v=wILX71HE3e0", "./youtube.mp4")
	require.NoError(t, err)

	err = d.Download(context.Background(), "https://www.youtube.com/watch?v=wILX71HE3e0", "./")
	require.NoError(t, err)
}
