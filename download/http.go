package download

import (
	"io"
	"net/http"
)

func NewHttpDownloader() *HttpDownloader {
	return &HttpDownloader{}
}

func (d *HttpDownloader) WithUrl(url string) *HttpDownloader {
	d.url = url
	return d
}

var _ Downloader = (*HttpDownloader)(nil)

type HttpDownloader struct {
	url string
}

func (d *HttpDownloader) Type() DownloadType {
	return HTTP
}

func (d *HttpDownloader) Download(path string) error {
	reader, err := d.Read()
	if err != nil {
		return err
	}

	return WriteToFile(reader, path)
}

func (d *HttpDownloader) Read() (io.Reader, error) {
	res, err := http.Get(d.url)
	if err != nil {
		return nil, err
	}

	return res.Body, nil
}
