package udpscan

import "time"

// ScanSign is the magic string used in UDP broadcast scanning protocol.
// ScanSign 是 UDP 广播扫描协议中使用的魔术字符串。
const ScanSign = "UDPSCAN"

// ScanPort is the default UDP port for scanning.
// ScanPort 是扫描使用的默认 UDP 端口。
const ScanPort = 42388

// Response represents the response from a host detection service.
// Response 表示主机检测服务的响应。
type Response struct {
	Name string    `json:"name"`
	Time time.Time `json:"time"`
}
