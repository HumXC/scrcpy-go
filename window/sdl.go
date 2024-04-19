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

type Window struct {
	window   *sdl.Window
	renderer *sdl.Renderer
	texture  *sdl.Texture
	width    int32
	height   int32
}

func (w *Window) Free() {
}

func (w *Window) RenderFrame(frame_ AVFrame) error {
	frame := frame_.P
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
	imageWidth := width_
	imageHeight := height_
	winWidth, winHeight := w.window.GetSize()
	winRatio := float32(winWidth) / float32(winHeight)
	imageRatio := float32(imageWidth) / float32(imageHeight)
	var x, y, width, height int32 = 0, 0, 0, 0
	if winRatio < imageRatio {
		scale := float64(winWidth) / float64(imageWidth)
		width = int32(float64(imageWidth) * scale)
		height = int32(float64(imageHeight) * scale)
		y = (winHeight - height) / 2
	} else {
		scale := float64(winHeight) / float64(imageHeight)
		width = int32(float64(imageWidth) * scale)
		height = int32(float64(imageHeight) * scale)
		x = (winWidth - width) / 2
	}

	// 计算图像在窗口中的位置
	destRect := &sdl.Rect{
		X: x,
		Y: y,
		W: width,
		H: height,
	}
	if err := w.renderer.Copy(w.texture, nil, destRect); err != nil {
		return err
	}
	w.renderer.Present()
	return nil
}
func sdlEvent(renderer *sdl.Renderer) (<-chan struct{}, <-chan struct{}) {
	quit := make(chan struct{}, 1)
	resize := make(chan struct{}, 1)
	go func() {
		for {
			for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
				switch event := event.(type) {
				case *sdl.QuitEvent:
					quit <- struct{}{}
					return
				case *sdl.WindowEvent:
					if event.Event == sdl.WINDOWEVENT_RESIZED {
						renderer.SetLogicalSize(event.Data1, event.Data2)
						resize <- struct{}{}
					}
				}
			}
		}
	}()
	return quit, resize
}

func InitWindow(title string, width, height uint32) (*Window, <-chan struct{}, <-chan struct{}, error) {
	w := &Window{}
	var err error
	defer func() {
		if err != nil {
			w.Free()
		}
	}()
	if err := sdl.Init(sdl.INIT_EVENTS); err != nil {
		panic(err)
	}
	w.window, err = sdl.CreateWindow(title, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, int32(width), int32(height), sdl.WINDOW_BORDERLESS|sdl.WINDOW_RESIZABLE|sdl.WINDOW_ALLOW_HIGHDPI)
	if err != nil {
		panic(err)
	}
	w.renderer, err = sdl.CreateRenderer(w.window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}
	quit, resize := sdlEvent(w.renderer)
	return w, quit, resize, nil
}
