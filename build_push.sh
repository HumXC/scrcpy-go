#!/usr/bin/env sh
go run ./build

adb push dist/scrcpy-go-server /data/local/tmp/ >/dev/null
adb shell "/data/local/tmp/scrcpy-go-server $@"
