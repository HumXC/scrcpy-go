一个 scrcpy 客户端实现，使用 quic 通信

该项目仅用于学习用途，实际意义不大，仅在 linux 系统测试。
使用 go 包装了 scrcpy 服务端以使用 quic 转发服务端的数据

客户端目前只实现了视频传输，在 40Mbps 的视频帧率加上 5G WIFI 体验较佳

### 安装环境

具体环境见 flake.nix，需要 ffmpeg, sdl2
如果你使用 nix，只需运行 `nix develop`

运行 `go run ./server/embeds/fetch.sh` 下载 scrcpy 服务端

### 使用方法

安卓设备连接主机并开启 adb
运行 `./build_push.sh` 将服务端推送到设备并运行

运行 `go run ./example/main.go` 启动客户端
