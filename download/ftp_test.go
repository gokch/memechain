package download

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFtpDownload(t *testing.T) {
	d, err := NewFtpDownloader("ftp://speedtest.tele2.net")
	require.NoError(t, err)
	err = d.Download(context.Background(), "512KB.zip", "./test.zip")
	require.NoError(t, err)
}
