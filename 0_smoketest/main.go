package main

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"time"
)

type Server struct {
	l net.Listener

	logger *slog.Logger
}

func NewServer(addr string) (*Server, error) {
	l, err := net.Listen("tcp", addr)

	if err != nil {
		return nil, err
	}

	s := &Server{
		l:      l,
		logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}

	return s, nil
}

func (s *Server) Close() error {
	return s.l.Close()
}

func (s *Server) ListenAndServe() error {
	for {
		c, err := s.l.Accept()

		if err != nil {
			s.logger.Error("connection error", "err", err.Error())
			continue
		}

		go s.handle(c)
	}
}

func (s *Server) handle(c net.Conn) {
	defer c.Close()

	c.SetDeadline(time.Now().Add(5 * time.Second))

	s.logger.Info("handling client", "addr", c.RemoteAddr())

	_, err := io.Copy(c, c)

	if err != nil {
		s.logger.Error(err.Error())
	}

	s.logger.Info("closing connection", "addr", c.RemoteAddr())
}

func main() {
	flagPort := flag.Int("p", 8000, "port to listen on")
	flag.Parse()

	var logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

	chanSignal := make(chan os.Signal, 1)
	signal.Notify(chanSignal, os.Interrupt)

	addr := fmt.Sprintf("0.0.0.0:%d", *flagPort)

	logger.Info("starting server", "addr", addr)
	srv, err := NewServer(addr)

	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	go func() {
		err = srv.ListenAndServe()

		if err != nil {
			logger.Error(err.Error())
		}
	}()

	<-chanSignal

	logger.Info("shutting down")
	os.Exit(0)
}
