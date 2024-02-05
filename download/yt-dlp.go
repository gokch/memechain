package download

import (
	"context"
	"errors"
	"io"

	"github.com/lrstanley/go-ytdlp"
)

func NewYtDlpDownloader() (*YtDlpDownloader, error) {
	_, err := ytdlp.Install(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	return &YtDlpDownloader{}, nil
}

var _ Downloader = (*YtDlpDownloader)(nil)

type YtDlpDownloader struct {
}

func (d *YtDlpDownloader) Type() DownloadType {
	return YTDLP
}

func (d *YtDlpDownloader) Download(ctx context.Context, remote, local string) error {
	if local[len(local)-1] == '/' {
		local += "%(extractor)s - %(title)s.%(ext)s"
	}
	cmd := ytdlp.New().Output(local)
	_, err := cmd.Run(context.Background(), remote)
	if err != nil {
		return err
	}
	return nil
}

func (d *YtDlpDownloader) Read(ctx context.Context, remote string) (io.Reader, error) {
	return nil, errors.New("not implemented")
}

func (d *YtDlpDownloader) Close() error {
	return nil
}
