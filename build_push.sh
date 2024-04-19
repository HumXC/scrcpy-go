#!/usr/bin/env sh
go run ./build

adb push dist/scrcpy-go-server /data/local/tmp/ >/dev/null
adb shell "QUIC_GO_LOG_LEVEL=INFO /data/local/tmp/scrcpy-go-server $@"
