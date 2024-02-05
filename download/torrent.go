package download

import (
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

func (d *TorrentDownloader) WithURI(uri string) *TorrentDownloader {
	d.uri = uri
	return d
}

var _ Downloader = (*TorrentDownloader)(nil)

type TorrentDownloader struct {
	client *torrent.Client
	uri    string
}

func (d *TorrentDownloader) Type() DownloadType {
	return TORRENT
}

func (d *TorrentDownloader) Download(path string) error {
	tor, err := d.client.AddMagnet(d.uri)
	if err != nil {
		return err
	}
	<-tor.GotInfo()
	tor.DownloadAll()
	return nil
}

func (d *TorrentDownloader) Read() (io.Reader, error) {
	return nil, errors.New("not implemented")
}
