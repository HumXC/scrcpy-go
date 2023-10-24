package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/HumXC/scrcpy-go/server"
	"github.com/shirou/gopsutil/process"
)

func Kill() {
	// 获取所有进程
	processes, err := process.Processes()
	if err != nil {
		panic(err)
	}

	// 遍历所有进程
	for _, proc := range processes {
		exe, err := proc.Exe()
		if err != nil {
			fmt.Println(exe)
			continue
		}
		fmt.Println(exe)
		// 以 Daemon 运行的进程名可能会是 "/data/local/tmp/shiroko (deleted)"
		exe = strings.TrimSuffix(exe, " (deleted)")

		if exe == server.ScrcpyServerPath {
			pid := proc.Pid
			// 检查 PID 是否是当前进程
			if pid != int32(os.Getpid()) {
				fmt.Printf("Killing process %d with name %s\n", pid, exe)
				_ = proc.Kill()
			}
		}
	}
}
func listenSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-c
		switch sig {
		case syscall.SIGINT:
			fmt.Println("\nReceived SIGINT, exiting...")
		case syscall.SIGTERM:
			fmt.Println("\nReceived SIGTERM, exiting...")
		}
		os.Exit(0)
	}()
}
func main() {
	opt, err := server.ParseOption(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// https://github.com/Genymobile/scrcpy/blob/master/doc/develop.md#execution
	opt.TunnelForward = true // https://github.com/Genymobile/scrcpy/blob/master/doc/develop.md#standalone-server
	opt.StayAwake = false
	opt.Cleanup = false
	opt.RawStream = true
	// opt.VideoEncoder = "OMX.google.h264.encoder"
	opt.Audio = false   // 暂时禁用，仅测试视频
	opt.Control = false // 暂时禁用，仅测试视频
	fmt.Println("Scrcpy Args:", opt.ToArgs())
	scrcpy := server.NewScrcpy(opt)
	err = scrcpy.Start()
	if err != nil {
		panic(err)
	}
	fmt.Println("Started")
	panic(server.StartHttpServer(scrcpy, ":8080"))
}
