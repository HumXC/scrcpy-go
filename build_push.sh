#!/usr/bin/env sh
go run ./build.go

adb push build/scrcpy-go-server /data/local/tmp/ >/dev/null
adb shell "/data/local/tmp/scrcpy-go-server $@"
