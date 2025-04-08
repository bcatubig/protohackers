package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"

	"github.com/bcatubig/protohackers/6_speed_daemon/server"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func main() {
	flagPort := flag.Int("p", 8000, "port to listen on")
	flag.Parse()

	addr := fmt.Sprintf("0.0.0.0:%d", *flagPort)
	srv := &server.Server{
		Addr: addr,
		Handler: server.HandlerFunc(func(c net.Conn) {
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
	srv.ListenAndServe()
}
