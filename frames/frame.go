package frames

import (
	"encoding/binary"
	"io"

	"github.com/HumXC/scrcpy-go/metas"
)

type Frame struct {
	Meta metas.Frame
	Data []byte
}

func Read(r io.Reader) (*Frame, error) {
	var f Frame
	err := f.Meta.LoadFrom(r)
	if err != nil {
		return nil, err
	}
	f.Data = make([]byte, f.Meta.PacketSize)
	err = binary.Read(r, binary.BigEndian, f.Data)
	if err != nil {
		return nil, err
	}
	return &f, nil
}
