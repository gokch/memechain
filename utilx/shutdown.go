package utilx

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
)

func HandleKillSig(handler func()) {
	sigChannel := make(chan os.Signal, 1)

	signal.Notify(sigChannel, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	go func() {
		for sig := range sigChannel {
			log.Info().Msgf("Receive signal %s, Shutting down...", sig)
			handler()
			os.Exit(1)
		}
	}()
}
