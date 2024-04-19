package main

import (
	"fmt"

	"github.com/HumXC/scrcpy-go/client"
	"github.com/HumXC/scrcpy-go/option"
	"github.com/HumXC/scrcpy-go/window"
)

func main() {
	// addr := "192.168.1.17:8080"
	addr := "192.168.1.18:8080"
	opt := option.Default()
	opt.Audio = false   // 暂时禁用，仅测试视频
	opt.Control = false // 暂时禁用，仅测试视频

	client := client.New(addr, opt)

	stream, err := client.Open()
	if err != nil {
		fmt.Println(err)
		return
	}
	video, err := stream.AsVideo()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer video.Close()
	fmt.Println(client.Name)
	fmt.Println(video.Codec)
	fmt.Println(video.Width, video.Height)
	sdl, quit, resize, err := window.InitWindow(client.Name, video.Width, video.Height)
	if err != nil {
		panic(err)
	}
	defer sdl.Free()

	dec := window.NewDecoder(video)
	defer dec.Free()
	defer func() {
		if dec.Error() != nil {
			fmt.Println(dec.Error())
		}
	}()
	frames := make(chan struct{})
	go func() {
		for {
			if dec.Next() {
				frames <- struct{}{}
			} else {
				close(frames)
				return
			}
		}
	}()
	var prevFrame *window.AVFrame
	for {
		select {
		case <-quit:
			return
		case <-resize:
			if prevFrame == nil {
				continue
			}
			_ = sdl.RenderFrame(*prevFrame)
		case _, ok := <-frames:
			if !ok {
				return
			}
			if prevFrame != nil {
				prevFrame.Free()
			}
			f := dec.Frame()
			prevFrame = &f
			_ = sdl.RenderFrame(f)
		}
	}
}
