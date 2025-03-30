package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func main() {
	flagPort := flag.Int("p", 8000, "port to listen on")
	flag.Parse()

	chanSignal := make(chan os.Signal, 1)
	signal.Notify(chanSignal, os.Interrupt)

	s := &http.Server{}

	addr := fmt.Sprintf("0.0.0.0:%d", *flagPort)
	srv, err := NewServer(addr)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	logger.Info("starting server", "addr", addr)
	go func() {
		srv.ListenAndServe()
	}()

	<-chanSignal
	logger.Info("shutting down server")
	if err := srv.Shutdown(context.Background()); err != nil {
		logger.Error(err.Error())
	}

	logger.Info("server exiting")
}
