package download

import (
	"context"
	"io"
	"net/http"
)

func NewHttpDownloader() (*HttpDownloader, error) {
	return &HttpDownloader{}, nil
}

var _ Downloader = (*HttpDownloader)(nil)

type HttpDownloader struct {
}

func (d *HttpDownloader) Type() DownloadType {
	return HTTP
}

func (d *HttpDownloader) Download(ctx context.Context, remote, local string) error {
	reader, err := d.Read(ctx, remote)
	if err != nil {
		return err
	}

	return WriteToFile(reader, local)
}

func (d *HttpDownloader) Read(ctx context.Context, remote string) (io.Reader, error) {
	res, err := http.Get(remote)
	if err != nil {
		return nil, err
	}

	return res.Body, nil
}

func (d *HttpDownloader) Close() error {
	return nil
}
