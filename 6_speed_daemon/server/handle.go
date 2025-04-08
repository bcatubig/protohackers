package server

type Handler interface {
	Serve(*Conn)
}

type HandlerFunc func(*Conn)

func (f HandlerFunc) Serve(c *Conn) {
	f(c)
}
