
#include <libavcodec/avcodec.h>
#include <libavformat/avformat.h>
#include <SDL2/SDL.h>
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