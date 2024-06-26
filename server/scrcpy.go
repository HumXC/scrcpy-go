package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/HumXC/scrcpy-go"
	"github.com/HumXC/scrcpy-go/logs"
)

const ServerVersion = "2.4"

type ScrcpyServer struct {
	bin       string
	opt       scrcpy.Options
	proc      *os.Process
	procState *os.ProcessState
	mu        sync.Mutex
}

func (s *ScrcpyServer) Option() scrcpy.Options {
	return s.opt
}

func New(opt scrcpy.Options, scrcpyBin string) *ScrcpyServer {
	return &ScrcpyServer{
		opt: opt,
		bin: scrcpyBin,
	}
}

func (s *ScrcpyServer) Start() (err error) {
	if s.proc != nil {
		return fmt.Errorf("scrcpy is running, pid: %d", s.proc.Pid)
	}

	cmd := exec.Command("app_process", append([]string{"/", "com.genymobile.scrcpy.Server", ServerVersion}, s.opt.ToArgs()...)...)
	cmd.Env = append(os.Environ(), "CLASSPATH="+s.bin)
	cmd.Stdout = logs.ScrcpyOutput
	cmd.Stderr = logs.ScrcpyOutput
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	defer func() {
		if err != nil {
			s.Stop()
		}
	}()
	err = cmd.Start()
	if err != nil {
		return
	}
	for {
		if cmd.Process != nil {
			if cmd.Err != nil {
				err = cmd.Err
				return
			} else {
				break
			}
		}
		time.Sleep(time.Millisecond)
	}
	s.proc = cmd.Process
	s.procState = cmd.ProcessState
	return nil
}
func (s *ScrcpyServer) TryOpen(timeout time.Duration) (io.ReadWriteCloser, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	for {
		conn, err := s.Open()
		if err == nil {
			return conn, nil
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("failed to try open socket: timeout %.2fs", timeout.Seconds())
		default:
			time.Sleep(time.Millisecond * 100)
		}
	}
}
func (s *ScrcpyServer) Open() (io.ReadWriteCloser, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	scid := ""
	if s.opt.Scid > 0 {
		scid += fmt.Sprintf("_%08d", s.opt.Scid)
	}
	conn, err := net.Dial("unix", "@scrcpy"+scid)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
func (s *ScrcpyServer) AutoOpen() (io.ReadWriteCloser, error) {
	if s.proc == nil || s.procState == nil || s.procState.Exited() {
		logs.GetLogger().Info("Scrcpy server not running, try to start")
		s.Stop()
		err := s.Start()
		if err != nil {
			return nil, err
		}
	}
	return s.TryOpen(3 * time.Second)
}
func (s *ScrcpyServer) Stop() error {
	if s.proc == nil {
		return nil
	}
	err := s.proc.Kill()
	if err != nil {
		return err
	}
	s.proc = nil
	s.procState = nil
	return nil
}
