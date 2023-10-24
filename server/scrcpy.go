package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path"
	"syscall"
	"time"

	"github.com/HumXC/scrcpy-go/server/binary"
)

const WORKDIR = "/data/local/tmp"
const ServerVersion = "2.1.1"

var ScrcpyServerPath = path.Join(WORKDIR, "scrcpy-server")

func IsExist(file string) bool {
	_, err := os.Stat(file)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
	}
	return false
}

type ScrcpyServer struct {
	opt     ScrcpyOptions
	proc    *os.Process
	audio   io.ReadCloser
	video   io.ReadCloser
	control io.ReadWriteCloser
}
type dropPool struct {
	puted map[io.ReadWriter]context.CancelFunc
}

func (p *dropPool) Put(rw io.ReadWriter) {
	ctx, cancel := context.WithCancel(context.Background())
	p.puted[rw] = cancel
	go func(ctx context.Context, rw io.ReadWriter) {
		buf := make([]byte, 1024)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				_, err := rw.Read(buf)
				if err != nil {
					return
				}
			}
		}
	}(ctx, rw)

}
func (p *dropPool) Get(c io.ReadWriter) {
	if cancel, ok := p.puted[c]; ok {
		cancel()
	}
	p.puted[c] = nil
}

func (s *ScrcpyServer) Option() ScrcpyOptions {
	return s.opt
}

func NewScrcpy(opt ScrcpyOptions) *ScrcpyServer {
	return &ScrcpyServer{
		opt: opt,
	}
}

// TODO: 实现广播，解决视频流等打开之后不被读取而造成的内存问题
func (s *ScrcpyServer) Start() error {
	if s.proc != nil {
		return fmt.Errorf("scrcpy is running, pid: %d", s.proc.Pid)
	}
	if !IsExist(ScrcpyServerPath) {
		err := os.WriteFile(ScrcpyServerPath, binary.ScrcpyServer, 0755)
		if err != nil {
			panic(err)
		}
	}
	cmd := exec.Command("app_process", append([]string{"/", "com.genymobile.scrcpy.Server", ServerVersion}, s.opt.ToArgs()...)...)
	cmd.Env = append(os.Environ(), "CLASSPATH="+ScrcpyServerPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	s.proc = cmd.Process
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	err := cmd.Start()
	if err != nil {
		return err
	}
	for {
		if cmd.Process != nil {
			if cmd.Err != nil {
				return cmd.Err
			} else {
				break
			}
		}
		time.Sleep(time.Millisecond)
	}

	if s.opt.Video {
		video, err := s.tryOpen(1 * time.Second)
		if err != nil {
			return err
		}
		s.video = video
		fmt.Println("opened scrcpy video")
	}
	if s.opt.Audio {
		audio, err := s.tryOpen(1 * time.Second)
		if err != nil {
			return err
		}
		s.audio = audio
		fmt.Println("opened scrcpy audio")
	}
	if s.opt.Control {
		s.control, err = s.tryOpen(1 * time.Second)
		if err != nil {
			return err
		}
		fmt.Println("opened scrcpy control")
	}
	return nil
}

func (s *ScrcpyServer) tryOpen(timeout time.Duration) (io.ReadWriteCloser, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("open failed, timeout")
		default:
			rw, err := s.open()
			if err != nil {
				time.Sleep(10 * time.Microsecond)
				continue
			}
			return rw, nil
		}
	}
}
func (s *ScrcpyServer) open() (io.ReadWriteCloser, error) {
	scid := "_"
	if s.opt.Scid > 0 {
		scid += fmt.Sprintf("%08d", s.opt.Scid)
	}
	conn, err := net.Dial("unix", "@scrcpy"+scid)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (s *ScrcpyServer) Video() (io.Reader, error) {
	if s.video == nil {
		return nil, fmt.Errorf("video not opened")
	}
	return s.video, nil
}

func (s *ScrcpyServer) Audio() (io.Reader, error) {
	if s.audio == nil {
		return nil, fmt.Errorf("audio not opened")
	}
	return s.audio, nil
}

func (s *ScrcpyServer) Control() (io.Writer, error) {
	if s.control == nil {
		return nil, fmt.Errorf("control not opened")
	}
	return s.control, nil
}

func (s *ScrcpyServer) Stop() error {
	if s.proc == nil {
		return nil
	}
	defer func() {
		s.video = nil
		s.audio = nil
		s.control = nil
	}()
	if s.video != nil {
		_ = s.video.Close()
	}
	if s.audio != nil {
		_ = s.audio.Close()
	}
	if s.control != nil {
		_ = s.control.Close()
	}
	err := s.proc.Kill()
	if err != nil {
		return err
	}
	s.proc = nil
	return nil
}
