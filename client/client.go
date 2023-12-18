package client

import (
	"context"
	"io"
	"time"

	"github.com/HumXC/scrcpy-go/tls"
	"github.com/quic-go/quic-go"
)

type ScrcpyClient struct {
	addr string
}

func (s *ScrcpyClient) Open() (io.ReadWriteCloser, error) {
	tlsConf, err := tls.Config()
	if err != nil {
		return nil, err
	}
	quicConfig := &quic.Config{}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	conn, err := quic.DialAddr(ctx, s.addr, tlsConf, quicConfig)
	if err != nil {
		return nil, err
	}
	ctx, cancel2 := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel2()
	stream, err := conn.AcceptStream(ctx)
	if err != nil {
		return nil, err
	}
	return stream, nil
}
func New(addr string) *ScrcpyClient {
	return &ScrcpyClient{addr: addr}
}
