package binary

import _ "embed"

//go:embed scrcpy-server-v2.1.1
var ScrcpyServer []byte

const ScrcpyServerVersion = "2.1.1"
