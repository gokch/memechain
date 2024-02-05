package download

import (
	"context"
	"errors"
	"io"

	"github.com/anacrolix/torrent"
)

// TODO
func NewTorrentDownloader() *TorrentDownloader {
	client, err := torrent.NewClient(nil)
	if err != nil {
		return nil
	}
	return &TorrentDownloader{
		client: client,
	}
}

var _ Downloader = (*TorrentDownloader)(nil)

type TorrentDownloader struct {
	client *torrent.Client
}

func (d *TorrentDownloader) Type() DownloadType {
	return TORRENT
}

func (d *TorrentDownloader) Download(ctx context.Context, remote, local string) error {
	tor, err := d.client.AddMagnet(remote)
	if err != nil {
		return err
	}
	<-tor.GotInfo()
	tor.DownloadAll()
	return nil
}

func (d *TorrentDownloader) Read(ctx context.Context, remote string) (io.Reader, error) {
	return nil, errors.New("not implemented")
}

func (d *TorrentDownloader) Close() error {
	d.client.Close()
	return nil
}
