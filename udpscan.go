package udpscan

import (
	"fmt"
	"net"
	"time"
)

// Version is the released version of this tool.
// Version 是本工具的发布版本号。
//
// ⚠️ Keep this in step with the git tag when releasing.
// ⚠️ 发版打标签时记得同步改这里。
const Version = "v0.0.1"

// ScanSign is the magic string used in UDP broadcast scanning protocol.
// ScanSign 是 UDP 广播扫描协议中使用的魔术字符串。
const ScanSign = "UDPSCAN"

// ScanPort is the default UDP port for scanning.
// ScanPort 是扫描使用的默认 UDP 端口。
const ScanPort = 42388

// DefaultSSHPort is the standard SSH port, omitted when composing commands.
// DefaultSSHPort 是标准的 SSH 端口，拼命令时会省略掉。
const DefaultSSHPort = 22

// MaxPacketSize is the largest payload a single UDP datagram can carry over IPv4.
// MaxPacketSize 是单个 IPv4 UDP 报文能承载的最大载荷，用作接收缓冲区大小。
//
// A buffer smaller than this truncates oversized replies, and a truncated reply
// fails to parse and gets dropped without a word — the hardest kind of failure.
// 缓冲区小于这个值会截断超长应答，而截断后的应答解析失败、被无声丢弃，最难排查。
const MaxPacketSize = 65507

// SafePayloadSize is the UDP payload that still fits one Ethernet frame unfragmented.
// SafePayloadSize 是仍能装进一个以太网帧、不触发 IP 分片的 UDP 载荷上限。
//
// 1500 (Ethernet MTU) - 20 (IPv4 header) - 8 (UDP header) = 1472, rounded down.
// 1500（以太网 MTU）- 20（IPv4 头）- 8（UDP 头）= 1472，取整留出余量。
// 超过之后报文会被分片，任一分片丢失整个报文即失效，且丢了不会有任何提示。
const SafePayloadSize = 1400

// Response represents the response from a host detection service.
// Response 表示主机检测服务的响应。
//
// The Username and SSHPort fields let the client compose a ready-to-run ssh command.
// Username 和 SSHPort 字段让客户端能拼出可直接执行的 ssh 命令。
type Response struct {
	Name string    `json:"name"`
	Time time.Time `json:"time"`
	// Username is the account to log in with, not necessarily the account running this service.
	// Username 是建议用来登录的账号名，未必是跑这个服务的账号（服务常以 root 运行）。
	Username string `json:"username,omitempty"`
	// SSHPort is the port sshd listens on, zero means the standard port.
	// SSHPort 是 sshd 实际监听的端口，零值表示标准端口。
	SSHPort int `json:"ssh_port,omitempty"`
}

// IsValidPort reports whether the value is a usable TCP/UDP port number.
// IsValidPort 判断给定的值是不是可用的 TCP/UDP 端口号。
//
// Zero is rejected here: it means "let the kernel choose" in some contexts, which is
// a caller-specific decision rather than a property of the number itself.
// 零在这里判为不可用：它在某些场景下表示"交给内核挑一个"，但那是调用方的语义，
// 不是这个数值本身的性质，所以留给调用方单独判断。
func IsValidPort(value int) bool {
	return value >= 1 && value <= 65535
}

// BroadcastAddress computes the IPv4 broadcast address of the network an address sits in.
// BroadcastAddress 算出给定地址所在网段的 IPv4 广播地址：网络位保持不变、主机位全部置 1。
//
// Returns nil when the address and mask are not a usable IPv4 pair, so callers can just
// skip that interface instead of sending probes to a bogus address.
// 地址与掩码不是可用的 IPv4 组合时返回 nil，调用方跳过该网卡即可，不会把探测包发到错地方。
func BroadcastAddress(ip net.IP, mask net.IPMask) net.IP {
	ipv4 := ip.To4()
	if ipv4 == nil {
		return nil // 广播是 IPv4 独有的概念，IPv6 没有广播
	}
	// 掩码可能是 IPv4-mapped 的十六字节形式，取后四字节才对得上
	if len(mask) == net.IPv6len {
		mask = mask[net.IPv6len-net.IPv4len:]
	}
	if len(mask) != net.IPv4len {
		return nil
	}
	result := make(net.IP, net.IPv4len)
	for i := range result {
		result[i] = ipv4[i] | ^mask[i]
	}
	return result
}

// Host is one discovered machine, seen from the client side.
// Host 表示客户端视角下发现的一台主机：服务端自报的内容，加上只有客户端才知道的 IP。
type Host struct {
	IP         string    `json:"ip"`
	Name       string    `json:"name"`
	Username   string    `json:"username,omitempty"`
	SSHPort    int       `json:"ssh_port,omitempty"`
	SSHCommand string    `json:"ssh_command,omitempty"`
	Time       time.Time `json:"time"`
}

// NewHost combines a response with the source IP it came from.
// NewHost 把服务端的应答和它的来源 IP 合成一条主机记录，顺带拼好 ssh 命令。
func NewHost(ip string, resp *Response) Host {
	return Host{
		IP:         ip,
		Name:       resp.Name,
		Username:   resp.Username,
		SSHPort:    resp.SSHPort,
		SSHCommand: resp.SSHCommand(ip),
		Time:       resp.Time,
	}
}

// ScanResult is the complete outcome of one scan, shaped for JSON output.
// ScanResult 是一次扫描的完整结果，形状是为 JSON 输出设计的。
//
// The overview fields matter most when Hosts is empty: they tell whether the scan
// found nothing because no host answered, or because it went out the wrong interface.
// 概览字段在 Hosts 为空时最有价值：能分辨"确实没有主机应答"还是"包压根发错了网卡"。
type ScanResult struct {
	Broadcasts     []string  `json:"broadcasts"`
	Port           int       `json:"port"`
	TimeoutSeconds int       `json:"timeout_seconds"`
	StartedAt      time.Time `json:"started_at"`
	WantedName     string    `json:"wanted_name,omitempty"`
	Count          int       `json:"count"`
	Hosts          []Host    `json:"hosts"`
}

// SSHCommand composes a ready-to-run ssh command from the response and the given IP.
// SSHCommand 用响应内容和给定的 IP 拼出可直接执行的 ssh 命令，拼不出时返回空串。
//
// The IP must come from the caller rather than from the response, because a host with
// several interfaces does not know which of its addresses actually reached the client.
// IP 必须由调用方传入、而不能由服务端自报：多网卡的主机并不知道自己是通过哪个地址被看到的，
// 只有客户端手里那个"实际收到应答的地址"才保证可达。
func (r *Response) SSHCommand(ip string) string {
	if r.Username == "" || ip == "" {
		return ""
	}
	if r.SSHPort != 0 && r.SSHPort != DefaultSSHPort {
		return fmt.Sprintf("ssh -p %d %s@%s", r.SSHPort, r.Username, ip)
	}
	return fmt.Sprintf("ssh %s@%s", r.Username, ip)
}
