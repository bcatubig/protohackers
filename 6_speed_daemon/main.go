package main

import (
	"flag"
	"fmt"
	"log/slog"
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
	}
	srv.ListenAndServe()
}
