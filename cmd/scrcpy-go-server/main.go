package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/HumXC/scrcpy-go/logs"
	"github.com/HumXC/scrcpy-go/server"
	"github.com/sevlyar/go-daemon"
	"github.com/shirou/gopsutil/process"
)

var (
	ip   string
	port string

	isDaemon bool
	isKill   bool
	isList   bool
)

func init() {
	flag.StringVar(&ip, "i", "0.0.0.0", "ip")
	flag.StringVar(&port, "p", "8080", "port")
	flag.BoolVar(&isDaemon, "d", false, "run with daemon mode")
	flag.BoolVar(&isKill, "k", false, "find and kill all process")
	flag.BoolVar(&isList, "l", false, "list all process")
	flag.Parse()
}

type Daemonize struct {
	IAmDaemon bool
	Pid       int
}

func Daemon(scid int) (pid int, err error) {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	_scid := ""
	if scid != -1 {
		_scid = fmt.Sprintf("-%08d", scid)
	}
	workdir := path.Dir(exe)
	cntxt := &daemon.Context{
		PidFileName: path.Join(workdir, fmt.Sprintf("scrcpy-go%s.pid", _scid)),
		PidFilePerm: 0644,
		LogFileName: path.Join(workdir, fmt.Sprintf("scrcpy-go%s.log", _scid)),
		LogFilePerm: 0640,
		WorkDir:     workdir,
		Umask:       027,
		Args:        os.Args,
	}
	defer cntxt.Release()
	proc, err := cntxt.Reborn()
	if err != nil {
		err = fmt.Errorf("unable to run: %s", err)
		return
	}
	if proc == nil {
		return
	}
	pid = proc.Pid
	return
}
func List() {
	target, err := os.Executable()
	if err != nil {
		panic(err)
	}
	for _, proc := range findProcess(target) {
		cmdl, _ := proc.Cmdline()
		fmt.Println(proc.Pid, cmdl)
	}
}
func findProcess(target string) []*process.Process {
	result := make([]*process.Process, 0, 1)
	processes, err := process.Processes()
	if err != nil {
		panic(err)
	}

	// 遍历所有进程
	for _, proc := range processes {
		exe, err := proc.Exe()
		if err != nil {
			continue
		}
		if strings.HasPrefix(exe, target) && proc.Pid != int32(os.Getpid()) {
			result = append(result, proc)
		}
	}
	return result
}
func Kill() {
	target, err := os.Executable()
	if err != nil {
		panic(err)
	}
	// 遍历所有进程
	for _, proc := range findProcess(target) {
		exe, _ := proc.Exe()
		fmt.Printf("Killing process %s pid is %d\n", exe, proc.Pid)
		_ = proc.Kill()
	}
}
func MDNS() {

}
func Command(scid int) {
	if isKill {
		Kill()
		os.Exit(0)
	}
	if isDaemon {
		pid, err := Daemon(scid)
		if err != nil {
			panic(err)
		}
		if pid != 0 {
			fmt.Printf("Run with daemon, pid is %d\n", pid)
			os.Exit(0)
		}
	}
	if isList {
		List()
		os.Exit(0)
	}
}
func main() {
	logger := logs.GetLogger()
	opt, err := server.ParseOption(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	Command(opt.Scid)

	// https://github.com/Genymobile/scrcpy/blob/master/doc/develop.md#execution
	opt.TunnelForward = true // https://github.com/Genymobile/scrcpy/blob/master/doc/develop.md#standalone-server
	// opt.StayAwake = false
	// opt.Cleanup = false
	opt.SendDummyByte = false
	opt.RawStream = false
	opt.VideoEncoder = "OMX.google.h264.encoder"
	opt.Audio = false   // 暂时禁用，仅测试视频
	opt.Control = false // 暂时禁用，仅测试视频
	logger.Info("Creating scrcpy", "args", opt.ToArgs())
	scrcpy := server.NewScrcpy(opt)
	defer scrcpy.Stop()
	server.RunServer(scrcpy, ip, port)
}
