package download

import (
	"context"
	"errors"
	"io"

	"github.com/anacrolix/torrent"
)

// TODO
func NewTorrentDownloader() (*TorrentDownloader, error) {
	client, err := torrent.NewClient(nil)
	if err != nil {
		return nil, err
	}
	return &TorrentDownloader{
		client: client,
	}, nil
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
	d.client.WaitAll()

	return nil
}

func (d *TorrentDownloader) Read(ctx context.Context, remote string) (io.Reader, error) {
	return nil, errors.New("not implemented")
}

func (d *TorrentDownloader) Close() error {
	d.client.Close()
	return nil
}
