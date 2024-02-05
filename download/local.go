package download

import (
	"context"
	"io"
	"os"
)

func NewLocalDownloader() (*LocalDownloader, error) {
	return &LocalDownloader{}, nil
}

var _ Downloader = (*LocalDownloader)(nil)

type LocalDownloader struct {
}

func (d *LocalDownloader) Type() DownloadType {
	return LOCAL
}

func (d *LocalDownloader) Download(ctx context.Context, remote, local string) error {
	reader, err := d.Read(ctx, remote)
	if err != nil {
		return err
	}
	return WriteToFile(reader, local)
}

func (d *LocalDownloader) Read(ctx context.Context, remote string) (io.Reader, error) {
	return os.Open(remote)
}

func (d *LocalDownloader) Close() error {
	return nil
}
