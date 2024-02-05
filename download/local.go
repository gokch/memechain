package download

import (
	"io"
	"os"
)

func NewLocalDownloader() *LocalDownloader {
	return &LocalDownloader{}
}

func (d *LocalDownloader) WithPath(path string) *LocalDownloader {
	d.path = path
	return d
}

var _ Downloader = (*LocalDownloader)(nil)

type LocalDownloader struct {
	path string
}

func (d *LocalDownloader) Type() DownloadType {
	return LOCAL
}

func (d *LocalDownloader) Download(path string) error {
	reader, err := d.Read()
	if err != nil {
		return err
	}
	return WriteToFile(reader, path)
}

func (d *LocalDownloader) Read() (io.Reader, error) {
	return os.Open(d.path)
}
