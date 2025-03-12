package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"math/big"
	"net"
	"os"
	"os/signal"
	"strconv"
)

type Server struct {
	l net.Listener

	logger *slog.Logger
}

type ServerOpt func(s *Server)

func WithListener(l net.Listener) ServerOpt {
	return func(s *Server) {
		s.l = l
	}
}

func WithLogger(l *slog.Logger) ServerOpt {
	return func(s *Server) {
		s.logger = l
	}
}

func NewServer(addr string, opts ...ServerOpt) (*Server, error) {
	s := &Server{}

	for _, opt := range opts {
		opt(s)
	}

	if s.l == nil {
		l, err := net.Listen("tcp", addr)

		if err != nil {
			return nil, err
		}

		s.l = l
	}

	if s.logger == nil {
		s.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}

	return s, nil
}

func (s *Server) ListenAndServe() error {
	for {
		c, err := s.l.Accept()
		if err != nil {
			continue
		}

		go s.handle(c)
	}
}

func (s *Server) handle(c net.Conn) {
	type request struct {
		Method string      `json:"method"`
		Number *DecimalInt `json:"number"`
	}

	type response struct {
		Method string `json:"method"`
		Prime  bool   `json:"prime"`
	}

	defer c.Close()

	reader := bufio.NewReader(c)

	for {
		data, err := reader.ReadBytes('\n')

		if err != nil {
			return
		}

		s.logger.Info("request received", "data", string(data))

		req := &request{}
		dec := json.NewDecoder(bytes.NewReader(data))

		err = dec.Decode(req)

		if err != nil {
			fmt.Println("error decoding json:", err.Error())
			c.Write(data)
			continue
		}

		if req.Method != "isPrime" || req.Number == nil {
			s.logger.Error("invalid request")
			c.Write(data)
			continue
		}

		if big.NewInt(int64(*req.Number)).ProbablyPrime(0) {
			s.logger.Info("is prime", "method", req.Method, "number", *req.Number)
			json.NewEncoder(c).Encode(&response{
				Method: "isPrime",
				Prime:  true,
			})
		} else {
			s.logger.Info("is not prime", "method", req.Method, "number", *req.Number)
			json.NewEncoder(c).Encode(&response{
				Method: "isPrime",
				Prime:  false,
			})
		}
	}

}

type DecimalInt int64

func (di DecimalInt) MarshalJSON() ([]byte, error) {
	return json.Marshal(int64(di))
}

func (di *DecimalInt) UnmarshalJSON(input []byte) error {
	r, err := strconv.ParseFloat(string(input), 64)

	if err != nil {
		return err
	}

	*di = DecimalInt(r)

	return nil
}

func main() {
	flagPort := flag.Int("p", 8000, "port to listen on")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	chanSignal := make(chan os.Signal, 1)
	signal.Notify(chanSignal, os.Interrupt)

	addr := fmt.Sprintf("0.0.0.0:%d", *flagPort)
	s, err := NewServer(addr, WithLogger(logger))

	if err != nil {
		logger.Error("error creating server", "error", err.Error())
		os.Exit(1)
	}

	go func() {
		err := s.ListenAndServe()

		if err != nil {
			logger.Error(err.Error())
		}
	}()

	<-chanSignal
	logger.Info("shutting down server")
}
