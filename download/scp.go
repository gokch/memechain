package download

import (
	"bytes"
	"context"
	"io"

	"github.com/bramvdbogaerde/go-scp"
)

func NewScpDownloader(url, path string) *ScpDownloader {
	client := scp.NewClient(url, nil)
	return &ScpDownloader{
		client: &client,
		path:   path,
	}
}

var _ Downloader = (*ScpDownloader)(nil)

type ScpDownloader struct {
	client *scp.Client
	path   string
}

func (d *ScpDownloader) Type() DownloadType {
	return SCP
}

func (d *ScpDownloader) Download(path string) error {
	reader, err := d.Read()
	if err != nil {
		return err
	}
	return WriteToFile(reader, path)
}

func (d *ScpDownloader) Read() (io.Reader, error) {
	err := d.client.Connect()
	if err != nil {
		return nil, err
	}
	defer d.client.Close()

	rw := bytes.NewBuffer(nil)
	err = d.client.CopyFromRemotePassThru(context.Background(), rw, d.path, nil)
	if err != nil {
		return nil, err
	}
	return rw, nil

}
