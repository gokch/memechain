package download

/*
import (
	"context"
	"io"

	"github.com/ipfs/boxo/tar"
	"github.com/ipfs/kubo/client/rpc"
	"github.com/multiformats/go-multiaddr"
)

func NewIPFSDownloader(url, cid string) *IPFSDownloader {
	ma, err := multiaddr.NewMultiaddr(url)
	if err != nil {
		return nil
	}
	api, err := rpc.NewApi(ma)
	if err != nil {
		return nil
	}
	return &IPFSDownloader{
		api: api,
		cid: cid,
	}
}

var _ Downloader = (*IPFSDownloader)(nil)

type IPFSDownloader struct {
	api *rpc.HttpApi
	cid string
}

func (d *IPFSDownloader) Download(path string) error {
	reader, err := d.Read()
	if err != nil {
		return err
	}
	extractor := &tar.Extractor{Path: path}
	return extractor.Extract(reader)
}

func (d *IPFSDownloader) Read() (io.Reader, error) {
	resp, err := d.api.Request("get", d.cid).Option("create", false).Send(context.Background())
	if err != nil {
		return nil, err
	}
	defer resp.Close()
	if resp.Error != nil {
		return nil, resp.Error
	}
	return resp.Output, nil
}
*/
