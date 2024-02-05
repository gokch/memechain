package download

import (
	"context"
	"errors"
	"io"

	"github.com/lrstanley/go-ytdlp"
)

func NewYtDlpDownloader() *YtDlpDownloader {
	ytdlp.MustInstall(context.Background(), nil)
	return &YtDlpDownloader{}
}

func (d *YtDlpDownloader) WithUrl(url string) *YtDlpDownloader {
	d.url = url
	return d
}

var _ Downloader = (*YtDlpDownloader)(nil)

type YtDlpDownloader struct {
	url string
}

func (d *YtDlpDownloader) Download(path string) error {
	if path[len(path)-1] == '/' {
		path += "%(extractor)s - %(title)s.%(ext)s"
	}
	cmd := ytdlp.New().Output(path)
	_, err := cmd.Run(context.Background(), d.url)
	if err != nil {
		return err
	}
	return nil
}

func (d *YtDlpDownloader) Read() (io.Reader, error) {
	return nil, errors.New("not implemented")
}
