package window

/*
#cgo LDFLAGS: -lSDL2 -lavformat

#include <SDL2/SDL.h>
#include <libavformat/avformat.h>
*/
import "C"
import (
	"fmt"
)

type Window struct {
	sdlWindow *C.SDL_Window
	title     *C.char
}

func (w *Window) Free() {
	C.SDL_DestroyWindow(w.sdlWindow)
	C.SDL_Quit()
}
func (w *Window) RenderFrame(frame *C.AVFrame) error {
	renderer := C.SDL_CreateRenderer(w.sdlWindow, -1, C.SDL_RENDERER_ACCELERATED)
	defer C.SDL_DestroyRenderer(renderer)
	if renderer == nil {
		return fmt.Errorf("SDL renderer create failed: %s", C.GoString(C.SDL_GetError()))
	}
	// 创建纹理
	texture := C.SDL_CreateTexture(renderer, C.SDL_PIXELFORMAT_YV12, C.SDL_TEXTUREACCESS_STREAMING, frame.width, frame.height)
	defer C.SDL_DestroyTexture(texture)
	if texture == nil {
		return fmt.Errorf("SDL texture create failed: %s", C.GoString(C.SDL_GetError()))
	}
	if C.SDL_UpdateYUVTexture(texture, nil,
		frame.data[0], frame.linesize[0],
		frame.data[1], frame.linesize[1],
		frame.data[2], frame.linesize[2]) < 0 {
		return fmt.Errorf("SDL texture update failed: %s", C.GoString(C.SDL_GetError()))
	}

	if C.SDL_RenderClear(renderer) < 0 {
		return fmt.Errorf("SDL renderer clear failed: %s", C.GoString(C.SDL_GetError()))
	}
	if C.SDL_RenderCopy(renderer, texture, nil, nil) < 0 {
		return fmt.Errorf("SDL renderer copy failed: %s", C.GoString(C.SDL_GetError()))
	}
	C.SDL_RenderPresent(renderer)
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
	if C.SDL_Init(C.SDL_INIT_VIDEO) < 0 {
		err = fmt.Errorf("SDL init failed: %s", C.GoString(C.SDL_GetError()))
		return nil, err
	}
	w.title = C.CString(title)
	w.sdlWindow = C.SDL_CreateWindow(w.title, C.SDL_WINDOWPOS_UNDEFINED, C.SDL_WINDOWPOS_UNDEFINED, C.int(width), C.int(height), C.SDL_WINDOW_SHOWN)
	if w.sdlWindow == nil {
		err = fmt.Errorf("SDL window creation failed: %s", C.GoString(C.SDL_GetError()))
		return nil, err
	}
	return w, nil
}
