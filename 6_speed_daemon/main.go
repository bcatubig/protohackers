package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/bcatubig/protohackers/6_speed_daemon/server"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func main() {
	chanSignal := make(chan os.Signal, 1)
	signal.Notify(chanSignal)

	mux := NewMux()
	mux.Register(64, server.HandlerFunc(func(c *server.Conn) {
		logger.Info("in handler 64")
		return
	}))

	srv := &server.Server{
		Handler: mux,
	}
	srv.WithLogger(logger)
	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			logger.Error(err.Error())
		}
	}()

	<-chanSignal

	srv.Shutdown(context.Background())
}
