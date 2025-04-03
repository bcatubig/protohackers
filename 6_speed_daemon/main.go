package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func main() {
	flagPort := flag.Int("p", 8000, "port to listen on")
	flag.Parse()

	chanSignal := make(chan os.Signal, 1)
	signal.Notify(chanSignal, os.Interrupt)

	addr := fmt.Sprintf("0.0.0.0:%d", *flagPort)

	svc := NewDispatcherService()

	srv := &Server{
		Addr:              addr,
		dispatcherService: svc,
	}

	logger.Info("starting server", "addr", addr)
	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			logger.Error(err.Error())
		}
	}()

	<-chanSignal
	logger.Info("shutting down server")
	if err := srv.Shutdown(); err != nil {
		logger.Error(err.Error())
	}

	logger.Info("server exiting")
}
