package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/bcatubig/protohackers/6_speed_daemon/server"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func main() {
	chanSignal := make(chan os.Signal, 1)
	signal.Notify(chanSignal)

	srv := &server.Server{
		Handler: server.HandlerFunc(func(c *server.Conn) {
			b := make([]byte, 256)
			for {
				n, err := c.Read(b)
				if err != nil {
					return
				}
				fmt.Println(string(b[:n]))

			}
		}),
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
