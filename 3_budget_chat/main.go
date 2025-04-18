package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"time"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func main() {
	flagPort := flag.Int("p", 8000, "port to listen on")
	flag.Parse()

	chanSignal := make(chan os.Signal, 1)
	signal.Notify(chanSignal, os.Interrupt)

	addr := fmt.Sprintf("0.0.0.0:%d", *flagPort)
	svr, err := NewServer(addr)

	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	logger.Info("starting server", "addr", addr)

	go func() {
		err := svr.ListenAndServe()
		if err != nil {
			logger.Error(err.Error())
		}
	}()

	<-chanSignal
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	svr.Shutdown(ctx)
	logger.Info("shutdown complete")
}
