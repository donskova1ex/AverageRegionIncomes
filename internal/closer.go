package internal

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type GracefulCloser struct {
	closingFunc []func() error
}

func NewGracefulCloser() *GracefulCloser {
	return &GracefulCloser{
		closingFunc: make([]func() error, 0),
	}
}

func (g *GracefulCloser) Add(closingFunc func() error) {
	g.closingFunc = append(g.closingFunc, closingFunc)
}

func (g *GracefulCloser) Run(ctx context.Context, log *slog.Logger) {
	wg := &sync.WaitGroup{}
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-signals:
		log.Info("closer signal received")
	case <-ctx.Done():
		log.Info("closer context cancelled")
	}
	wg.Add(len(g.closingFunc))
	for _, closingFunc := range g.closingFunc {
		go func(closingFunc func() error) {
			defer wg.Done()
			if err := closingFunc(); err != nil {
				log.Error("error during closing function")
			}
		}(closingFunc)
	}
	wg.Wait()
	log.Info("everything closed correctly")
}
