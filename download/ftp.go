package download

import (
	"context"
	"io"

	"github.com/jlaffaye/ftp"
)

func NewFtpDownloader(host string) *FtpDownloader {
	client, err := ftp.Dial(host)
	if err != nil {
		return nil
	}
	return &FtpDownloader{
		client: client,
	}
}

var _ Downloader = (*FtpDownloader)(nil)

type FtpDownloader struct {
	client *ftp.ServerConn
}

func (d *FtpDownloader) Type() DownloadType {
	return FTP
}

func (d *FtpDownloader) Download(ctx context.Context, remote, local string) error {
	reader, err := d.Read(ctx, remote)
	if err != nil {
		return err
	}

	return WriteToFile(reader, local)
}

func (d *FtpDownloader) Read(ctx context.Context, remote string) (io.Reader, error) {
	res, err := d.client.Retr(remote)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *FtpDownloader) Close() error {
	return d.client.Quit()
}
