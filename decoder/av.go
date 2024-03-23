package decoder

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
*/
import "C"
import (
	"errors"
	"fmt"
	"io"
	"unsafe"
)

type Decoder struct {
	id     C.int
	packet *C.AVPacket
	frame  *C.AVFrame
	err    error

	formatCtx        *C.AVFormatContext
	avioCtx          *C.AVIOContext
	codecCtx         *C.AVCodecContext
	buffer           *C.uchar
	videoStreamIndex C.int
	ret              C.int
}

func (s *Decoder) Next() bool {
	if s.err != nil {
		return false
	}
	if C.av_read_frame(s.formatCtx, s.packet) < 0 {
		return false
	}
	if s.packet.stream_index != s.videoStreamIndex {
		return false
	}
	// 发送数据到解码器
	s.ret = C.avcodec_send_packet(s.codecCtx, s.packet)
	if s.ret < 0 {
		s.err = errors.New("Error sending a packet for decoding")
		return false
	}
	s.ret = C.avcodec_receive_frame(s.codecCtx, s.frame)
	if s.ret == C.AVERROR_(C.EAGAIN) || s.ret == C.AVERROR_EOF {
		return false
	} else if s.ret < 0 {
		s.err = errors.New("Error during decoding")
		return false
	}
	return true
}
func (s *Decoder) Error() error {
	return s.err
}
func (s *Decoder) Frame() *C.AVFrame {
	return s.frame
}
func (s *Decoder) Free() {
	C.free(unsafe.Pointer(&s.id))
	C.av_packet_free(&s.packet)
	C.av_frame_free(&s.frame)
	C.av_packet_unref(s.packet)

	C.avformat_close_input(&s.formatCtx)
	C.avio_context_free(&s.avioCtx)
	C.free(unsafe.Pointer(s.buffer))
	C.avcodec_free_context(&s.codecCtx)
	C.free(unsafe.Pointer(&s.ret))
	C.free(unsafe.Pointer(&s.videoStreamIndex))
}
func New(input io.Reader) *Decoder {
	s := &Decoder{}
	id := C.int(0)
	for i := range readers {
		if i == id {
			id++
		}
	}
	readers[id] = input
	s.formatCtx = C.avformat_alloc_context()
	const bufferSize = 4096
	s.buffer = (*C.uchar)(C.malloc(bufferSize))

	s.avioCtx = C.avio_alloc_context_with_go_IO(
		s.buffer,
		bufferSize,
		0,
		unsafe.Pointer(&s.id),
	)
	s.formatCtx.pb = s.avioCtx

	if C.avformat_open_input(&s.formatCtx, nil, nil, nil) != 0 {
		panic("Cannot open input file")
	}

	// 获取流信息
	if C.avformat_find_stream_info(s.formatCtx, nil) < 0 {
		panic("Cannot find stream information")
	}
	// 查找视频流
	s.videoStreamIndex = C.av_find_best_stream(s.formatCtx, C.AVMEDIA_TYPE_VIDEO, -1, -1, nil, 0)
	if s.videoStreamIndex < 0 {
		panic("Cannot find video stream in input file")
	}
	// 获取视频流解码器
	codec := C.avcodec_find_decoder(C.at_streams(s.formatCtx.streams, s.videoStreamIndex).codecpar.codec_id)
	if codec == nil {
		panic("Unsupported codec")
	}

	s.codecCtx = C.avcodec_alloc_context3(codec)
	if s.codecCtx == nil {
		panic("Could not allocate codec context")
	}

	// 将流参数拷贝到解码器上下文cket
	if C.avcodec_parameters_to_context(s.codecCtx, C.at_streams(s.formatCtx.streams, s.videoStreamIndex).codecpar) < 0 {
		fmt.Println("Failed to copy codec parameters to decoder context")
	}

	// 打开解码器
	if C.avcodec_open2(s.codecCtx, codec, nil) < 0 {
		panic("Could not open codec")
	}

	// 分配AVPacket和AVFrame
	s.packet = C.av_packet_alloc()
	s.frame = C.av_frame_alloc()

	return s
}

var readers map[C.int]io.Reader = make(map[C.int]io.Reader)

//export readPacket
func readPacket(id unsafe.Pointer, buf *C.uint8_t, bufSize C.int) C.int {
	id_ := *(*C.int)(id)
	reader := readers[id_]
	if reader == nil {
		return 0
	}
	goBuffer := make([]byte, int(bufSize))
	n, err := reader.Read(goBuffer)

	if err != nil && err != io.EOF {
		return C.AVERROR_EOF
	}

	if n > 0 {
		C.memcpy(unsafe.Pointer(buf), unsafe.Pointer(&goBuffer[0]), C.size_t(n))
	}

	if err == io.EOF {
		return C.AVERROR_EOF
	}

	return C.int(n)
}
