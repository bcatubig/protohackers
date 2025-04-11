package main

import (
	"log/slog"
	"os"
)

var logger *slog.Logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

type server struct{}

func main() {
	s := &Server{}
	s.ListenAndServe()
}
