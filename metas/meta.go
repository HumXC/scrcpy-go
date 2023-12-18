package metas

import (
	"encoding/binary"
	"io"

	"github.com/HumXC/scrcpy-go/codecs"
)

type Device struct {
	Name string
}

func (d *Device) LoadFrom(r io.Reader) {
	name := make([]byte, 64)
	binary.Read(r, binary.BigEndian, &name)
	d.Name = string(name)
}

func (d *Device) ToBytes() []byte {
	data := []byte(d.Name)

	if len(data) < 64 {
		padding := make([]byte, 64-len(data))
		data = append(data, padding...)
	} else {
		data = data[:64]
	}
	return data
}

type VideoCodec struct {
	Codec  codecs.Codec
	Width  uint32
	Height uint32
}

func (d *VideoCodec) LoadFrom(r io.Reader) error {
	data := make([]byte, 12)
	err := binary.Read(r, binary.BigEndian, data)
	if err != nil {
		return err
	}
	d.Codec = codecs.FromId(binary.BigEndian.Uint32(data[0:4]))
	d.Width = binary.BigEndian.Uint32(data[4:8])
	d.Height = binary.BigEndian.Uint32(data[8:12])
	return nil
}
func (d *VideoCodec) ToBytes() []byte {
	data := make([]byte, 12)
	binary.BigEndian.PutUint32(data[0:4], d.Codec.Id)
	binary.BigEndian.PutUint32(data[4:8], d.Width)
	binary.BigEndian.PutUint32(data[8:12], d.Height)
	return data
}

type AudioCodec struct {
	Codec codecs.Codec
}

func (d *AudioCodec) LoadFrom(r io.Reader) error {
	var data uint32
	err := binary.Read(r, binary.BigEndian, &data)
	if err != nil {
		return err
	}
	d.Codec = codecs.FromId(data)
	return nil
}
func (d *AudioCodec) ToBytes() []byte {
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, d.Codec.Id)
	return data
}

const (
	PACKET_FLAG_CONFIG    = uint64(1) << 63
	PACKET_FLAG_KEY_FRAME = uint64(1) << 62
)

type Frame struct {
	ConfigPacketFlag bool
	KeyFrameFlag     bool
	PTS              uint64
	PacketSize       uint32
}

func (f *Frame) LoadFrom(r io.Reader) error {
	pts := make([]byte, 8)
	err := binary.Read(r, binary.BigEndian, pts)
	if err != nil {
		return err
	}
	f.ConfigPacketFlag = (binary.BigEndian.Uint64(pts) & PACKET_FLAG_CONFIG) != 0
	f.KeyFrameFlag = (binary.BigEndian.Uint64(pts) & PACKET_FLAG_KEY_FRAME) != 0
	f.PTS = binary.BigEndian.Uint64(pts) &^ (PACKET_FLAG_CONFIG | PACKET_FLAG_KEY_FRAME)
	err = binary.Read(r, binary.BigEndian, &f.PacketSize)
	if err != nil {
		return err
	}
	return nil
}
