package window

/*
#cgo LDFLAGS: -lSDL2 -lavformat

#include <libavformat/avformat.h>
*/
import "C"
import (
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

func init() {

}

type Window struct {
	window   *sdl.Window
	renderer *sdl.Renderer
	texture  *sdl.Texture
	width    int32
	height   int32
}

func (w *Window) Free() {
}

func (w *Window) RenderFrame(frame *C.AVFrame) error {
	yPlane := (*uint8)(unsafe.Pointer(frame.data[0]))
	uPlane := (*uint8)(unsafe.Pointer(frame.data[1]))
	vPlane := (*uint8)(unsafe.Pointer(frame.data[2]))
	yLinesize := int(frame.linesize[0])
	uLinesize := int(frame.linesize[1])
	vLinesize := int(frame.linesize[2])

	yLength := yLinesize * int(frame.height)
	uLength := uLinesize * int(frame.height/2)
	vLength := vLinesize * int(frame.height/2)

	ySlice := (*[1 << 30]byte)(unsafe.Pointer(yPlane))[:yLength:yLength]
	uSlice := (*[1 << 30]byte)(unsafe.Pointer(uPlane))[:uLength:uLength]
	vSlice := (*[1 << 30]byte)(unsafe.Pointer(vPlane))[:vLength:vLength]

	width_ := int32(frame.width)
	height_ := int32(frame.height)

	if width_ != w.width || height_ != w.height {
		w.width = width_
		w.height = height_
		if w.texture != nil {
			if err := w.texture.Destroy(); err != nil {
				return err
			}
		}
		w.texture = nil
	}
	if w.texture == nil {
		var err error
		w.texture, err = w.renderer.CreateTexture(sdl.PIXELFORMAT_IYUV, sdl.TEXTUREACCESS_STREAMING, w.width, w.height)
		if err != nil {
			return err
		}
	}
	if err := w.texture.UpdateYUV(nil,
		ySlice, yLinesize,
		uSlice, uLinesize,
		vSlice, vLinesize,
	); err != nil {
		return err
	}
	if err := w.renderer.Copy(w.texture, nil, nil); err != nil {
		return err
	}
	w.renderer.Present()
	return nil
}
func NewWindow(title string, width, height uint32) (*Window, error) {
	w := &Window{}
	var err error
	defer func() {
		if err != nil {
			w.Free()
		}
	}()
	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		panic(err)
	}
	w.window, err = sdl.CreateWindow(title, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, int32(width), int32(height), sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	w.renderer, err = sdl.CreateRenderer(w.window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}

	return w, nil
}
