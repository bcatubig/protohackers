package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/bcatubig/protohackers/pkg/tcp"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func main() {
	chanSignal := make(chan os.Signal, 1)
	signal.Notify(chanSignal)

	srv := &tcp.Server{
		Handler: &Mux{},
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
