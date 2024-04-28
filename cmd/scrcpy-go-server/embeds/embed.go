package embeds

import _ "embed"

//go:embed scrcpy-server-v2.4
var ScrcpyServer []byte

// 修改版本号应该同步修改 go:embed 行
// 以及 fetch.sh 的版本号
const ScrcpyServerVersion = "2.4"
