package download

import (
	"context"
	"fmt"
	"io"
	"os"
)

func New(downloadType DownloadType, host string) (Downloader, error) {
	switch downloadType {
	case LOCAL:
		return NewLocalDownloader()
	case HTTP:
		return NewHttpDownloader()
	case TORRENT:
		return NewTorrentDownloader()
	case FTP:
		return NewFtpDownloader(host)
	case YTDLP:
		return NewYtDlpDownloader()
	default:
		return nil, fmt.Errorf("unknown download type: %d", downloadType)
	}
}

type Downloader interface {
	Type() DownloadType
	Download(ctx context.Context, remote, local string) error
	Read(ctx context.Context, remote string) (io.Reader, error)
	Close() error
}

type DownloadType uint8

const (
	// common protocol [ 0 - 31 ]
	LOCAL DownloadType = 0x00
	HTTP  DownloadType = 0x01
	FTP   DownloadType = 0x02
	SCP   DownloadType = 0x03

	// p2p protocol [ 32 - 63 ]
	TORRENT DownloadType = 0x20
	IPFS    DownloadType = 0x21
	SWARM   DownloadType = 0x22

	// use specific downloader [ 64 - 127 ]
	YTDLP DownloadType = 0x40

	// use custom downloader [ 128 - 254 ]

	// unknown protocol - 255
	UNKNOWN DownloadType = 0xFF
)

func FromString(s string) DownloadType {
	switch s {
	case "LOCAL":
		return LOCAL
	case "HTTP":
		return HTTP
	case "FTP":
		return FTP
	case "SCP":
		return SCP
	case "TORRENT":
		return TORRENT
	case "IPFS":
		return IPFS
	case "SWARM":
		return SWARM
	case "YTDLP":
		return YTDLP
	default:
		return UNKNOWN
	}
}

func (t DownloadType) ToString() string {
	switch t {
	case LOCAL:
		return "LOCAL"
	case HTTP:
		return "HTTP"
	case FTP:
		return "FTP"
	case SCP:
		return "SCP"
	case TORRENT:
		return "TORRENT"
	case IPFS:
		return "IPFS"
	case SWARM:
		return "SWARM"
	case YTDLP:
		return "YTDLP"
	default:
		return "UNKNOWN"
	}
}

func WriteToFile(reader io.Reader, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, reader)
	if err != nil {
		return err
	}
	return nil
}
