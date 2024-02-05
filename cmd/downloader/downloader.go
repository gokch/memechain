package main

import (
	"context"
	"os"
	"time"

	"github.com/gokch/memechain/download"
	"github.com/gokch/memechain/utilx"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "downloader",
		Short: "file downloader cli using multiple protocols",
		Run:   rootRun,
	}

	rootPath   string
	timeout    int64
	workerSize int64

	downloadType string
	host         string
	remotes      []string
	locals       []string
)

func init() {
	fs := rootCmd.PersistentFlags()
	fs.StringVarP(&rootPath, "rootpath", "p", "./", "root path")
	fs.Int64VarP(&timeout, "timeout", "t", 0, "timeout seconds, 0 is no timeout")
	fs.Int64VarP(&workerSize, "worker", "w", 1, "worker size")

	fs.StringVar(&downloadType, "type", "t", "download type")
	fs.StringVar(&host, "host", "h", "host address")
	fs.StringArrayVarP(&remotes, "remotes", "r", []string{}, "remote path or url")
	fs.StringArrayVarP(&locals, "locals", "l", []string{}, "local path")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func rootRun(cmd *cobra.Command, args []string) {
	var ctx context.Context
	var cancel context.CancelFunc

	ctx = context.Background()
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	}
	if cancel != nil {
		defer cancel()
	}
	log.Info().Msg("start cli")

	downloader, err := download.New(download.FromString(downloadType), host)
	if err != nil {
		log.Error().Err(err).Msg("new downloader")
	}
	for i := 0; i < len(remotes); i++ {
		downloader.Download(ctx, remotes[i], locals[i])
	}

	utilx.HandleKillSig(func() {
		err = downloader.Close()
		if err != nil {
			log.Error().Err(err).Msg("close downloader")
		}
	})
}
