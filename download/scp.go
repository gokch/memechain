package download

import (
	"bytes"
	"context"
	"io"

	"github.com/bramvdbogaerde/go-scp"
)

func NewScpDownloader(url string) *ScpDownloader {
	client := scp.NewClient(url, nil)
	return &ScpDownloader{
		client: &client,
	}
}

var _ Downloader = (*ScpDownloader)(nil)

type ScpDownloader struct {
	client *scp.Client
}

func (d *ScpDownloader) Type() DownloadType {
	return SCP
}

func (d *ScpDownloader) Download(ctx context.Context, remote, local string) error {
	reader, err := d.Read(ctx, remote)
	if err != nil {
		return err
	}
	return WriteToFile(reader, local)
}

func (d *ScpDownloader) Read(ctx context.Context, remote string) (io.Reader, error) {
	err := d.client.Connect()
	if err != nil {
		return nil, err
	}
	defer d.client.Close()

	rw := bytes.NewBuffer(nil)
	err = d.client.CopyFromRemotePassThru(ctx, rw, remote, nil)
	if err != nil {
		return nil, err
	}
	return rw, nil
}

func (d *ScpDownloader) Close() error {
	d.client.Close()
	return nil
}
