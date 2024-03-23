package server

import (
	"context"
	"io"
	"net"

	"github.com/HumXC/scrcpy-go/logs"
	"github.com/HumXC/scrcpy-go/tls"
	"github.com/quic-go/quic-go"
)

type quicServer struct {
	scrcpy *ScrcpyServer
}

func (s *quicServer) Run(ip, port string) error {
	logger := logs.GetLogger()
	addr, err := net.ResolveUDPAddr("udp4", ip+":"+port)
	if err != nil {
		return err
	}
	udpConn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return err
	}
	tr := quic.Transport{
		Conn: udpConn,
	}
	tlsConf, err := tls.Config()
	if err != nil {
		return err
	}
	quicConf := &quic.Config{}
	ln, err := tr.Listen(tlsConf, quicConf)
	if err != nil {
		return err
	}
	logger.Info("QUIC server started", "ip", ip, "port", port)
	for {
		conn, err := ln.Accept(context.Background())
		if err != nil {
			logger.Error("QUIC connect error", "msg", err)
		}
		go s.handler(conn)
	}
}
func (s *quicServer) handler(conn quic.Connection) {
	logger := logs.GetLogger()
	logger.Info("QUIC connect success", "remote", conn.RemoteAddr())
	stream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		logger.Error("QUIC open stream error", "msg", err)
		return
	}
	scrcpySocket, err := s.scrcpy.AutoOpen()
	if err != nil {
		logger.Error("Scrcpy open error", "msg", err)
		return
	}
	defer scrcpySocket.Close()
	defer logger.Info("QUIC connect closed", "remote", conn.RemoteAddr())

	_, err = io.Copy(stream, scrcpySocket)
	if err != nil {
		defer logger.Error("QUIC io error", "msg", err)
		return
	}
}
func NewQUIC(scrcpy *ScrcpyServer) *quicServer {
	return &quicServer{
		scrcpy: scrcpy,
	}
}
