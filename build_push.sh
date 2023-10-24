#!/usr/bin/env sh
export GOOS=linux
export GOARCH=arm64
go build -o ./build/scrcpy-go-server ./cmd/scrcpy-go-server
adb push build/scrcpy-go-server /data/local/tmp/ >/dev/null
adb shell "/data/local/tmp/scrcpy-go-server $@"
