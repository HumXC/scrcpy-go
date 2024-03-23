package client

import (
	"context"
	"encoding/binary"
	"errors"
	"io"
	"time"

	"github.com/HumXC/scrcpy-go/codecs"
	"github.com/HumXC/scrcpy-go/metas"
	"github.com/HumXC/scrcpy-go/option"
	"github.com/HumXC/scrcpy-go/tls"
	"github.com/quic-go/quic-go"
)

type VideoStream struct {
	io.ReadCloser
	Codec  codecs.Codec
	Width  uint32
	Height uint32
}
type AudioStream struct {
	io.ReadCloser
	Codec codecs.Codec
}
type ControlStream = io.ReadWriteCloser

type Stream struct {
	rwc io.ReadWriteCloser
	opt option.Options
}

func (s *Stream) AsVideo() (*VideoStream, error) {
	if !s.opt.Video {
		return nil, errors.New("video stream is not allowed")
	}
	v := VideoStream{ReadCloser: s.rwc}
	if !s.opt.RawStream && s.opt.SendCodecMeta {
		codec := metas.VideoCodec{}
		err := codec.LoadFrom(s.rwc)
		if err != nil {
			return nil, err
		}
		v.Codec = codec.Codec
		v.Width = codec.Width
		v.Height = codec.Height
	}
	return &v, nil
}
func (s *Stream) AsAudio() (*AudioStream, error) {
	if !s.opt.Audio {
		return nil, errors.New("audio stream is not allowed")
	}
	a := AudioStream{ReadCloser: s.rwc}
	if !s.opt.RawStream && s.opt.SendCodecMeta {
		codec := metas.AudioCodec{}
		err := codec.LoadFrom(s.rwc)
		if err != nil {
			return nil, err
		}
		a.Codec = codec.Codec
	}
	return &a, nil
}
func (s *Stream) AsControl() (ControlStream, error) {
	if !s.opt.Control {
		return nil, errors.New("control stream is not allowed")
	}
	return s.rwc, nil
}

type ScrcpyClient struct {
	Addr           string
	Name           string
	Opt            option.Options
	isNotFirstOpem bool
}

func (s *ScrcpyClient) Open() (*Stream, error) {
	tlsConf, err := tls.Config()
	if err != nil {
		return nil, err
	}
	quicConfig := &quic.Config{}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	conn, err := quic.DialAddr(ctx, s.Addr, tlsConf, quicConfig)
	if err != nil {
		return nil, err
	}
	ctx, cancel2 := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel2()
	stream, err := conn.AcceptStream(ctx)
	if err != nil {
		return nil, err
	}
	if s.isNotFirstOpem {
		return &Stream{rwc: stream, opt: s.Opt}, nil
	}
	s.isNotFirstOpem = true
	if s.Opt.RawStream {
		return &Stream{rwc: stream, opt: s.Opt}, nil
	}
	if s.Opt.SendDummyByte {
		var dummyByte byte = 0
		err := binary.Read(stream, binary.BigEndian, &dummyByte)
		if err != nil {
			stream.Close()
			return nil, err
		}
	}
	if s.Opt.SendDeviceMeta {
		device := metas.Device{}
		err = device.LoadFrom(stream)
		if err != nil {
			stream.Close()
			return nil, err
		}
		s.Name = device.Name
	}
	return &Stream{rwc: stream, opt: s.Opt}, nil
}
func New(addr string, opt option.Options) *ScrcpyClient {
	return &ScrcpyClient{Addr: addr, Opt: opt}
}
