package download

import (
	"io"

	"github.com/jlaffaye/ftp"
)

func NewFtpDownloader(address string) *FtpDownloader {
	client, err := ftp.Dial(address)
	if err != nil {
		return nil
	}
	return &FtpDownloader{
		client: client,
	}
}

var _ Downloader = (*FtpDownloader)(nil)

type FtpDownloader struct {
	client     *ftp.ServerConn
	remotePath string
}

func (d *FtpDownloader) Type() DownloadType {
	return FTP
}

func (d *FtpDownloader) Download(path string) error {
	reader, err := d.Read()
	if err != nil {
		return err
	}

	return WriteToFile(reader, path)
}

func (d *FtpDownloader) Read() (io.Reader, error) {
	res, err := d.client.Retr(d.remotePath)
	if err != nil {
		return nil, err
	}
	return res, nil
}
