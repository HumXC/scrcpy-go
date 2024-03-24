
#include <libavcodec/avcodec.h>
#include <libavformat/avformat.h>

int readPacket(void *opaque, uint8_t *buf, int bufSize);

AVStream *at_streams(AVStream **streams, int index)
{
    return streams[index];
}

int AVERROR_(int code)
{
    return AVERROR(code);
}
AVIOContext *avio_alloc_context_with_go_IO(
    unsigned char *buffer,
    int buffer_size,
    int write_flag,
    void *opaque)
{
    return avio_alloc_context(buffer, buffer_size, write_flag, opaque, readPacket, NULL, NULL);
}

static enum AVPixelFormat get_hw_format(AVCodecContext *ctx,
                                        const enum AVPixelFormat *pix_fmts)
{
    const enum AVPixelFormat *p;

    for (p = pix_fmts; *p != -1; p++)
    {
        fprintf(stderr, " format %d\n", *p);

        if (*p == AV_PIX_FMT_VAAPI)
            return *p;
    }

    fprintf(stderr, "Failed to get HW surface format\n");
    return AV_PIX_FMT_NONE;
}