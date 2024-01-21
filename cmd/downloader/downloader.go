package main

import (
	"context"
	"os"
	"time"

	"github.com/gokch/memechain/utilx"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use: "client",
		Run: rootRun,
	}

	rootPath   string
	timeout    int64
	workerSize int64
	expireSec  int64

	peerIds []string
	cids    []string
	paths   []string
)

func init() {
	fs := rootCmd.PersistentFlags()
	fs.StringVarP(&rootPath, "rootpath", "r", "./", "root path")
	fs.Int64VarP(&timeout, "timeout", "t", 0, "timeout seconds, 0 is no timeout")
	fs.Int64VarP(&workerSize, "worker", "w", 1, "worker size")
	fs.Int64VarP(&expireSec, "expire", "e", 600, "expire seconds")

	fs.StringArrayVar(&peerIds, "peers", []string{}, "connect peer id")
	fs.StringArrayVarP(&cids, "cids", "c", []string{}, "download cid")
	fs.StringArrayVarP(&paths, "paths", "p", []string{}, "download path per cid")
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

	utilx.HandleKillSig(func() {
		// TODO
	})

}
