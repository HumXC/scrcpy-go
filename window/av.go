package window

/*
#cgo LDFLAGS: -lavcodec -lavformat -lavutil

#include <libavcodec/avcodec.h>
#include <libavformat/avformat.h>

AVStream *at_streams(AVStream **streams, int index);
int AVERROR_(int code);
AVIOContext *avio_alloc_context_with_go_IO(
    unsigned char *buffer,
    int buffer_size,
    int write_flag,
    void *opaque);
void set_hw_frames_ctx(AVCodecContext *ctx, AVBufferRef *hw_frames_ctx);
*/
import "C"
import (
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"unsafe"
)

func avError(err C.int) error {
	if err == 0 {
		return nil
	}
	var errbuf [C.AV_ERROR_MAX_STRING_SIZE]byte
	cErrbuf := (*C.char)(unsafe.Pointer(&errbuf[0]))

	if C.av_strerror(err, cErrbuf, C.AV_ERROR_MAX_STRING_SIZE) < 0 {
		return errors.New("unknown error")
	}
	return errors.New(strings.Trim(string(errbuf[:]), "\x00"))
}

type Decoder struct {
	r      io.Reader
	err    error
	id     C.int
	packet *C.AVPacket
	frame  *C.AVFrame

	formatCtx        *C.AVFormatContext
	avioCtx          *C.AVIOContext
	codecCtx         *C.AVCodecContext
	buffer           *C.AVBufferRef
	videoStreamIndex C.int
	ret              C.int
	mu               sync.Mutex
}
type AVFrame struct {
	P *C.AVFrame
}

func (f AVFrame) Free() {
	C.av_frame_unref(f.P)
	C.av_frame_free(&f.P)
}
func (s *Decoder) Free() {
	C.avcodec_free_context(&s.codecCtx)
	C.avformat_free_context(s.formatCtx)
	C.avio_context_free(&s.avioCtx)
	C.av_frame_free(&s.frame)
	C.av_packet_free(&s.packet)
	// FIXME 此处 buffer 无法 free
	// C.av_buffer_unref(&s.buffer)
	delete(decoders, s.id)
}

func (s *Decoder) Next() bool {
	if s.err != nil {
		return false
	}
	C.av_packet_unref(s.packet)
	frame := C.av_frame_alloc()
	if err := avError(C.av_read_frame(s.formatCtx, s.packet)); err != nil {
		s.err = fmt.Errorf("reading frame failed: %w", err)
		return false
	}
	if s.packet.stream_index != s.videoStreamIndex {
		return false
	}
	s.ret = C.avcodec_send_packet(s.codecCtx, s.packet)
	if s.ret < 0 {
		s.err = fmt.Errorf("sending a packet for decoding: %w", avError(s.ret))
		return false
	}
	s.ret = C.avcodec_receive_frame(s.codecCtx, frame)
	if s.ret == C.AVERROR_(C.EAGAIN) || s.ret == C.AVERROR_EOF {
		return false
	} else if s.ret < 0 {
		s.err = fmt.Errorf("during decoding: %w", avError(s.ret))
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.frame != nil {
		s.err = errors.New("frame already exists")
		return false
	}
	s.frame = frame
	return true
}
func (s *Decoder) Error() error {
	return s.err
}
func (s *Decoder) Frame() AVFrame {
	s.mu.Lock()
	defer s.mu.Unlock()
	f := AVFrame{P: s.frame}
	s.frame = nil
	return f
}

func NewDecoder(input io.Reader) *Decoder {
	const bufferSize = 4096 * 1024
	s := &Decoder{
		r: input,
	}
	id := C.int(0)
	for i := range decoders {
		if i == id {
			id++
		}
	}
	decoders[id] = s
	s.formatCtx = C.avformat_alloc_context()
	s.buffer = C.av_buffer_alloc(bufferSize)

	s.avioCtx = C.avio_alloc_context_with_go_IO(
		s.buffer.data,
		bufferSize,
		0,
		unsafe.Pointer(&s.id),
	)
	s.formatCtx.pb = s.avioCtx
	if err := avError(C.avformat_open_input(&s.formatCtx, nil, nil, nil)); err != nil {
		s.err = fmt.Errorf("cannot open input: %w", err)
		return s
	}
	s.videoStreamIndex = C.av_find_best_stream(s.formatCtx, C.AVMEDIA_TYPE_VIDEO, -1, -1, nil, 0)
	if s.videoStreamIndex < 0 {
		err := avError(s.videoStreamIndex)
		s.err = fmt.Errorf("cannot find video stream in input file: %w", err)
		return s
	}
	params := C.at_streams(s.formatCtx.streams, s.videoStreamIndex).codecpar
	codec := C.avcodec_find_decoder(params.codec_id)
	if codec == nil {
		s.err = errors.New("unsupported codec")
		return s
	}

	s.codecCtx = C.avcodec_alloc_context3(codec)
	if s.codecCtx == nil {
		s.err = errors.New("could not allocate codec context")
		return s
	}
	if err := avError(C.avcodec_parameters_to_context(s.codecCtx, params)); err != nil {
		s.err = fmt.Errorf("failed to copy codec parameters to decoder context: %w", err)
		return s
	}
	if err := avError(C.avcodec_open2(s.codecCtx, codec, nil)); err != nil {
		s.err = fmt.Errorf("could not open codec; %w", err)
		return s
	}
	s.packet = C.av_packet_alloc()

	return s
}

var decoders map[C.int]*Decoder = make(map[C.int]*Decoder)

//export readPacket
func readPacket(id unsafe.Pointer, buf *C.uint8_t, bufSize C.int) C.int {
	id_ := *(*C.int)(id)
	decoder := decoders[id_]
	if decoder == nil {
		return 0
	}
	buffer := (*[1 << 30]byte)(unsafe.Pointer(buf))[:int(bufSize):int(bufSize)]
	n, err := decoder.r.Read(buffer)
	decoder.err = err
	if err == io.EOF || n == 0 {
		return C.AVERROR_EOF
	}
	if err != nil {
		return C.AVERROR_UNKNOWN
	}
	return C.int(n)
}
