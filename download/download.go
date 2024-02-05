package download

import (
	"io"
	"os"
)

type Downloader interface {
	Type() DownloadType
	Download(path string) error
	Read() (io.Reader, error)
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
