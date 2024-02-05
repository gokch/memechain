package download

import (
	"bytes"
	"context"
	"io"

	"github.com/bramvdbogaerde/go-scp"
)

func NewSCPDownloader(url, path string) *SCPDownloader {
	client := scp.NewClient(url, nil)
	return &SCPDownloader{
		client: &client,
		path:   path,
	}
}

var _ Downloader = (*SCPDownloader)(nil)

type SCPDownloader struct {
	client *scp.Client
	path   string
}

func (d *SCPDownloader) Download(path string) error {
	reader, err := d.Read()
	if err != nil {
		return err
	}
	return WriteToFile(reader, path)
}

func (d *SCPDownloader) Read() (io.Reader, error) {
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
