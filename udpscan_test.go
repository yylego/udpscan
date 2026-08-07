package udpscan

import (
	"encoding/json"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBroadcastAddress(t *testing.T) {
	t.Run("按位算出各种掩码下的广播地址", func(t *testing.T) {
		items := []struct {
			mask   string // CIDR 写法，便于一眼看懂
			ip     string
			expect string
		}{
			{"/24", "192.168.1.10", "192.168.1.255"},
			{"/16", "172.16.5.3", "172.16.255.255"},
			{"/8", "10.1.2.3", "10.255.255.255"},
			// 非字节对齐的掩码最能验证位运算，整字节的掩码就算写错也可能碰巧对
			{"/25", "192.168.1.10", "192.168.1.127"},
			{"/25", "192.168.1.200", "192.168.1.255"},
			{"/30", "10.0.0.5", "10.0.0.7"},
			{"/32", "10.0.0.5", "10.0.0.5"}, // 单机掩码，广播就是它自己
		}
		for _, item := range items {
			_, ipNet, err := net.ParseCIDR(item.ip + item.mask)
			require.NoError(t, err)
			result := BroadcastAddress(net.ParseIP(item.ip), ipNet.Mask)
			require.Equal(t, item.expect, result.String(), "输入 %s%s", item.ip, item.mask)
		}
	})

	t.Run("接受十六字节的IPv4映射掩码", func(t *testing.T) {
		mask := make(net.IPMask, net.IPv6len)
		copy(mask[net.IPv6len-net.IPv4len:], net.IPv4Mask(255, 255, 255, 0))
		result := BroadcastAddress(net.ParseIP("192.168.1.10"), mask)
		require.Equal(t, "192.168.1.255", result.String())
	})

	t.Run("IPv6地址没有广播概念", func(t *testing.T) {
		_, ipNet, err := net.ParseCIDR("fd00::1/64")
		require.NoError(t, err)
		require.Nil(t, BroadcastAddress(ipNet.IP, ipNet.Mask))
	})

	t.Run("掩码长度不对时返回空", func(t *testing.T) {
		require.Nil(t, BroadcastAddress(net.ParseIP("192.168.1.10"), net.IPMask{255, 255}))
	})
}

func TestResponseFitsInOneFrame(t *testing.T) {
	// 典型内容的应答必须能装进一个不分片的以太网帧。
	// 将来往 Response 加字段时，这条会挡住"加着加着就超了"——超了会被分片，
	// 丢一个分片整包作废，且两端都不会报错。
	resp := Response{
		Name:     "a-fairly-long-hostname-used-in-testing",
		Time:     time.Now(),
		Username: "a-fairly-long-username-used-in-testing",
		SSHPort:  22,
	}
	data, err := json.Marshal(resp)
	require.NoError(t, err)
	require.Less(t, len(data), SafePayloadSize, "应答包超过单帧安全载荷，会被分片")
}

func TestIsValidPort(t *testing.T) {
	t.Run("合法范围", func(t *testing.T) {
		require.True(t, IsValidPort(1))
		require.True(t, IsValidPort(22))
		require.True(t, IsValidPort(ScanPort))
		require.True(t, IsValidPort(65535))
	})
	t.Run("零不算合法端口", func(t *testing.T) {
		// 0 在有些场景表示"由内核随机挑"，但那是调用方的语义，不是数值本身合法
		require.False(t, IsValidPort(0))
	})
	t.Run("越界与负数", func(t *testing.T) {
		require.False(t, IsValidPort(-1))
		require.False(t, IsValidPort(65536))
		require.False(t, IsValidPort(99999))
	})
}

func TestPacketSizeConstants(t *testing.T) {
	// 接收缓冲区必须能装下任何合法的 UDP 报文，否则超长应答会被截断后静默丢弃
	require.Equal(t, 65507, MaxPacketSize)
	require.Less(t, SafePayloadSize, MaxPacketSize)
}

func TestVersionMatchesTagShape(t *testing.T) {
	// 版本号要跟 git 标签同形，发版时两边才对得上
	require.Regexp(t, `^v\d+\.\d+\.\d+$`, Version)
}

func TestScanSign(t *testing.T) {
	require.Equal(t, "UDPSCAN", ScanSign)
	require.NotEmpty(t, ScanSign)
}

func TestScanPort(t *testing.T) {
	require.Equal(t, 42388, ScanPort)
	require.Positive(t, ScanPort)
}

func TestResponse(t *testing.T) {
	now := time.Now()
	resp := Response{Name: "test-host", Time: now}
	require.Equal(t, "test-host", resp.Name)
	require.Equal(t, now, resp.Time)
}

func TestSSHCommand(t *testing.T) {
	t.Run("标准端口时省略-p", func(t *testing.T) {
		resp := Response{Name: "n1", Username: "yyle88", SSHPort: DefaultSSHPort}
		require.Equal(t, "ssh yyle88@10.42.0.5", resp.SSHCommand("10.42.0.5"))
	})

	t.Run("端口为零时按标准端口处理", func(t *testing.T) {
		// 老版本服务端不带这个字段，反序列化后就是零值
		resp := Response{Name: "n1", Username: "yyle88"}
		require.Equal(t, "ssh yyle88@10.42.0.5", resp.SSHCommand("10.42.0.5"))
	})

	t.Run("非标准端口时带上-p", func(t *testing.T) {
		resp := Response{Name: "n1", Username: "yyle88", SSHPort: 2222}
		require.Equal(t, "ssh -p 2222 yyle88@10.42.0.5", resp.SSHCommand("10.42.0.5"))
	})

	t.Run("没有用户名就拼不出命令", func(t *testing.T) {
		resp := Response{Name: "n1"}
		require.Empty(t, resp.SSHCommand("10.42.0.5"))
	})

	t.Run("没有IP也拼不出命令", func(t *testing.T) {
		resp := Response{Name: "n1", Username: "yyle88"}
		require.Empty(t, resp.SSHCommand(""))
	})
}

func TestResponseCompatibleWithOldPayload(t *testing.T) {
	// 老版本服务端只发 name 和 time，新客户端必须能正常解析、且不误报登录方式
	const oldPayload = `{"name":"old-host","time":"2025-01-01T00:00:00Z"}`

	var resp Response
	require.NoError(t, json.Unmarshal([]byte(oldPayload), &resp))
	require.Equal(t, "old-host", resp.Name)
	require.Empty(t, resp.Username)
	require.Zero(t, resp.SSHPort)
	require.Empty(t, resp.SSHCommand("10.42.0.5"))
}

func TestResponseOmitsEmptyFields(t *testing.T) {
	// 没配用户名时不该往报文里塞空字段，保持跟老报文一致的形状
	data, err := json.Marshal(Response{Name: "n1", Time: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)})
	require.NoError(t, err)
	require.NotContains(t, string(data), "username")
	require.NotContains(t, string(data), "ssh_port")
}

func TestResponseJSON(t *testing.T) {
	resp := Response{Name: "xiaozhixiang", Time: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)}

	data, err := json.Marshal(resp)
	require.NoError(t, err)
	require.Contains(t, string(data), `"name":"xiaozhixiang"`)

	var decoded Response
	require.NoError(t, json.Unmarshal(data, &decoded))
	require.Equal(t, resp.Name, decoded.Name)
	require.Equal(t, resp.Time, decoded.Time)
}
