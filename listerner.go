package grpcmock

import (
	"net"
	"sync"

	"go.nhat.io/grpcmock/must"
)

var _ net.Listener = (*listener)(nil)

type listener struct {
	upstream   net.Listener
	signal     chan struct{}
	sendSignal sync.Once
}

// Accept waits for and returns the next connection to the listener.
func (l *listener) Accept() (net.Conn, error) {
	l.sendSignal.Do(func() {
		close(l.signal)
	})

	return l.upstream.Accept()
}

// Close closes the listener.
func (l *listener) Close() error {
	return l.upstream.Close()
}

// Addr returns the listener's network address.
func (l *listener) Addr() net.Addr {
	return l.upstream.Addr()
}

func newListenerByAddr(addr string) func() (net.Listener, func() error) {
	return func() (net.Listener, func() error) {
		l, err := net.Listen("tcp", addr)
		must.NotFail(err)

		return l, l.Close
	}
}

func newListenerWithReadySignal(upstream net.Listener) (net.Listener, <-chan struct{}) {
	signal := make(chan struct{})

	return &listener{
		upstream: upstream,
		signal:   signal,
	}, signal
}
