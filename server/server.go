package server

import (
	"context"
	"fmt"
	"io"
	"net"

	"github.com/HumXC/scrcpy-go/logs"
	"github.com/HumXC/scrcpy-go/tls"
	"github.com/quic-go/quic-go"
)

type quicServer struct {
	scrcpy *ScrcpyServer
}

func (s *quicServer) Run(ip, port string) {
	logger := logs.GetLogger()
	addr, err := net.ResolveUDPAddr("udp4", ip+":"+port)
	if err != nil {
		fmt.Println(err)
		return
	}
	udpConn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		fmt.Println(err)
		return
	}
	tr := quic.Transport{
		Conn: udpConn,
	}
	tlsConf, err := tls.Config()
	if err != nil {
		fmt.Println(err)
		return
	}
	quicConf := &quic.Config{}
	ln, err := tr.Listen(tlsConf, quicConf)
	if err != nil {
		fmt.Println(err)
		return
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
	io.Copy(stream, scrcpySocket)
	logger.Info("QUIC connect closed", "remote", conn.RemoteAddr())
}
func RunServer(scrcpy *ScrcpyServer, ip, port string) {
	s := &quicServer{
		scrcpy: scrcpy,
	}
	s.Run(ip, port)
}
