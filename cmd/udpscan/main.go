package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/user"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/yylego/must"
	"github.com/yylego/udpscan"
)

var (
	name       string
	username   string
	sshPort    int
	port       int // server 用：本机监听端口
	broadcast  string
	wantedName string
	asJSON     bool
	targetPort int // client 用：对方监听的端口
	sourcePort int // client 用：本机发包与收应答的端口
	timeout    int
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "udpscan",
		Short:   "UDP 局域网主机发现工具",
		Version: udpscan.Version, // 顺带提供 --version
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "显示版本号与协议标识",
		Run:   runVersion,
	}

	serverCmd := &cobra.Command{
		Use:     "server",
		Short:   "启动服务端，监听并回复自己的昵称",
		PreRunE: validateServerFlags,
		Run:     runServer,
	}
	serverCmd.Flags().StringVarP(&name, "name", "n", "", "主机昵称 (必填)")
	serverCmd.Flags().StringVarP(&username, "username", "u", "", "供对方 ssh 登录的用户名 (留空则取当前账号)")
	serverCmd.Flags().IntVar(&sshPort, "ssh-port", udpscan.DefaultSSHPort, "本机 sshd 监听的端口")
	serverCmd.Flags().IntVarP(&port, "port", "p", udpscan.ScanPort, "监听端口")
	must.Done(serverCmd.MarkFlagRequired("name"))

	clientCmd := &cobra.Command{
		Use:     "client",
		Short:   "扫描局域网内的主机",
		PreRunE: validateClientFlags,
		Run:     runClient,
	}
	clientCmd.Flags().StringVarP(&broadcast, "broadcast", "b", "", "广播地址 (留空则自动枚举所有网卡)")
	clientCmd.Flags().StringVarP(&wantedName, "name", "n", "", "只保留昵称匹配的主机 (可能匹配到多台)")
	clientCmd.Flags().BoolVar(&asJSON, "json", false, "输出 JSON, 便于脚本处理")
	clientCmd.Flags().IntVarP(&targetPort, "port", "p", udpscan.ScanPort, "【对方】监听的端口, 探测包发往这里")
	clientCmd.Flags().IntVar(&sourcePort, "source-port", 0, "【本机】发包与收应答用的端口 (0=随机, 固定后防火墙只需放行一个口)")
	clientCmd.Flags().IntVarP(&timeout, "timeout", "t", 3, "扫描超时(秒)")

	rootCmd.AddCommand(serverCmd, clientCmd, versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// validateServerFlags 在启动服务前把参数核一遍。
//
// 无效参数不该等到运行时才以奇怪的方式显形：--ssh-port 传 0 会让对方拿到一条
// 省掉了 -p 的错误 ssh 命令，昵称留空则会让扫描方看到一台没有名字的机器。
func validateServerFlags(cmd *cobra.Command, args []string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("--name 不能为空")
	}
	if !udpscan.IsValidPort(port) {
		return fmt.Errorf("--port 必须在 1-65535 之间, 当前是 %d", port)
	}
	if !udpscan.IsValidPort(sshPort) {
		return fmt.Errorf("--ssh-port 必须在 1-65535 之间, 当前是 %d", sshPort)
	}
	return nil
}

// validateClientFlags 在扫描前把参数核一遍。
//
// ⚠️ --timeout 传 0 或负数最坑：扫描会立刻结束并报告"共发现 0 台主机"——
// 那是个天天都见的正常输出，没人会想到是参数写错了，于是错怪到网络头上。
func validateClientFlags(cmd *cobra.Command, args []string) error {
	if !udpscan.IsValidPort(targetPort) {
		return fmt.Errorf("--port 必须在 1-65535 之间, 当前是 %d", targetPort)
	}
	// 0 在这里是有意义的：表示由内核随机挑一个端口
	if sourcePort != 0 && !udpscan.IsValidPort(sourcePort) {
		return fmt.Errorf("--source-port 必须是 0(随机) 或 1-65535, 当前是 %d", sourcePort)
	}
	if timeout <= 0 {
		return fmt.Errorf("--timeout 必须大于 0 秒, 当前是 %d", timeout)
	}
	return nil
}

// runVersion 输出版本号和协议标识。
//
// 两端能不能对上话，取决于【协议魔术字】而不是版本号 —— 魔术字不一致时请求照发、
// 服务端照收，只是不匹配就直接丢弃，两边都不会打印任何异常。所以这里把魔术字也打出来：
// 两台机器各跑一次一比，立刻能判断是不是协议不兼容，比抓包快得多。
func runVersion(cmd *cobra.Command, args []string) {
	fmt.Printf("udpscan     %s\n", udpscan.Version)
	fmt.Printf("协议魔术字  %s\n", udpscan.ScanSign)
	fmt.Printf("默认端口    %d\n", udpscan.ScanPort)
	fmt.Printf("Go 版本     %s\n", runtime.Version())
}

// resolveUsername 定出要告诉对方的登录用户名：显式配置优先，否则取当前用户。
//
// 服务通常以 root 常驻（macOS 的 LaunchDaemon、Linux 的 init.d 都是），而 root
// 几乎不会是对方想登录的账号，所以这种情况下宁可留空、也不报一个误导的名字。
func resolveUsername(configured string) string {
	if configured != "" {
		return configured
	}
	userInfo, err := user.Current()
	if err != nil {
		return ""
	}
	if userInfo.Username == "root" {
		return ""
	}
	return userInfo.Username
}

func runServer(cmd *cobra.Command, args []string) {
	listenAddress := &net.UDPAddr{IP: net.IPv4(0, 0, 0, 0), Port: port}
	conn, err := net.ListenUDP("udp", listenAddress)
	if err != nil {
		fmt.Println("启动失败:", err)
		os.Exit(1)
	}
	defer func() { _ = conn.Close() }()

	loginUsername := resolveUsername(username)

	fmt.Printf("服务端启动 [%s] 监听端口 %d\n", name, port)
	if loginUsername == "" {
		fmt.Println("提示: 未定出登录用户名, 对方扫到本机后无法直接得到 ssh 命令, 建议加 --username 指定")
	} else {
		fmt.Printf("对外通告的登录方式: %s@本机 (sshd 端口 %d)\n", loginUsername, sshPort)
	}

	// 先试算一次应答包多大：超过安全载荷就会被 IP 分片，任一分片丢失整个报文即作废，
	// 而且丢了不会有任何提示。在这里说一声，比日后排查"这台机器怎么扫不到"省事得多。
	probe := udpscan.Response{Name: name, Time: time.Now(), Username: loginUsername, SSHPort: sshPort}
	if data, err := json.Marshal(probe); err == nil && len(data) > udpscan.SafePayloadSize {
		fmt.Printf("提示: 应答包约 %d 字节, 已超过 %d 的安全载荷, 可能因分片而丢失, 建议缩短昵称\n",
			len(data), udpscan.SafePayloadSize)
	}

	buf := make([]byte, 1024)
	for {
		n, clientAddress, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("读取错误:", err)
			continue
		}

		msg := string(buf[:n])
		if msg == udpscan.ScanSign {
			resp := udpscan.Response{
				Name:     name,
				Time:     time.Now(),
				Username: loginUsername,
				SSHPort:  sshPort,
			}
			// 序列化失败意味着 Response 结构本身有问题，是确定性的代码错误 ——
			// 每次都会失败，静默跳过就成了"服务在跑却永不应答"的假象，那是最难查的一类。
			// 宁可当场崩掉：服务管理器会重启并留下日志，问题立刻显形。
			data, err := json.Marshal(resp)
			must.Done(err)

			// 发送失败则是偶发的环境问题（对方已走、缓冲区满），记一笔继续服务即可，不该崩。
			if _, err := conn.WriteToUDP(data, clientAddress); err != nil {
				fmt.Fprintf(os.Stderr, "回复 %s 失败: %v\n", clientAddress.IP, err)
				continue
			}
			fmt.Printf("回复 %s\n", clientAddress.IP)
		}
	}
}

// resolveBroadcastAddresses 枚举本机每张网卡各自的 IPv4 广播地址。
//
// ⭐ 点对点接口（VPN、代理建的 tun / utun）不带 net.FlagBroadcast 标志，
// 判断这个标志就天然把它们滤掉了，不用按名字硬编码去排除谁。
func resolveBroadcastAddresses() []net.IP {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil
	}

	var results []net.IP
	for _, item := range interfaces {
		if item.Flags&net.FlagUp == 0 {
			continue // 网卡没启用
		}
		if item.Flags&net.FlagLoopback != 0 {
			continue // 回环口上没有别的主机
		}
		if item.Flags&net.FlagBroadcast == 0 {
			continue // 不支持广播，点对点隧道就落在这里
		}

		addresses, err := item.Addrs()
		if err != nil {
			continue
		}
		for _, entry := range addresses {
			ipNet, ok := entry.(*net.IPNet)
			if !ok {
				continue
			}
			if result := udpscan.BroadcastAddress(ipNet.IP, ipNet.Mask); result != nil {
				results = append(results, result)
			}
		}
	}
	return results
}

// resolveScanTargets 定出要往哪些地址发探测包：显式指定了就只用它，否则逐网卡广播。
func resolveScanTargets(configured string) []net.IP {
	if configured != "" {
		single := net.ParseIP(configured)
		if single == nil {
			fmt.Fprintln(os.Stderr, "无效的广播地址:", configured)
			os.Exit(1)
		}
		return []net.IP{single}
	}
	return resolveBroadcastAddresses()
}

func runClient(cmd *cobra.Command, args []string) {
	targets := resolveScanTargets(broadcast)
	if len(targets) == 0 {
		fmt.Fprintln(os.Stderr, "没有找到支持广播的网卡, 可以用 -b 手动指定广播地址")
		os.Exit(1)
	}

	// 监听本地端口用于接收应答。默认 0 让内核随机挑一个；
	// 指定固定端口的意义在于防火墙：应答包的目的端口就是这里，
	// 端口固定下来，放行规则才能精确到一个口，而不必放开整个临时端口范围。
	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(0, 0, 0, 0), Port: sourcePort})
	if err != nil {
		fmt.Fprintln(os.Stderr, "监听失败:", err)
		// 指定了固定端口时最常见的原因是被占用（比如同时跑了两个扫描）。
		// 这里【不自动退回随机端口】：那样防火墙规则会突然不匹配，
		// 变成"有时扫得到有时扫不到"的诡异现象，比直接失败难查得多。
		if sourcePort != 0 {
			fmt.Fprintf(os.Stderr, "提示: 端口 %d 可能已被占用, 去掉 --source-port 即可改用随机端口\n", sourcePort)
		}
		os.Exit(1)
	}
	defer func() { _ = listener.Close() }()

	// 逐个广播地址发探测包：某个网卡发不出去不影响其它网卡
	var reached []string
	for _, target := range targets {
		broadcastAddress := &net.UDPAddr{IP: target, Port: targetPort}
		if _, err := listener.WriteToUDP([]byte(udpscan.ScanSign), broadcastAddress); err != nil {
			fmt.Fprintf(os.Stderr, "发送到 %s 失败: %v\n", target, err)
			continue
		}
		reached = append(reached, target.String())
	}
	if len(reached) == 0 {
		fmt.Fprintln(os.Stderr, "所有广播地址都发送失败")
		os.Exit(1)
	}

	startedAt := time.Now()
	must.Done(listener.SetReadDeadline(startedAt.Add(time.Duration(timeout) * time.Second)))

	// JSON 模式下这些进度信息不能进 stdout，否则会把 JSON 冲脏
	if !asJSON {
		fmt.Printf("扫描中... (广播: %s, 端口: %d, 超时: %ds)\n", strings.Join(reached, " "), targetPort, timeout)
		fmt.Println("--------------------")
	}

	// 初始化成空切片而不是 nil：JSON 里要输出 [] 而不是 null，调用方才能直接遍历
	hosts := make([]udpscan.Host, 0)
	seen := make(map[string]bool)
	// 按单个 UDP 报文的上限开缓冲区：小了会把超长应答截断，
	// 而截断后的应答解析不出来、被下面的 continue 无声丢掉，那台主机就凭空消失了
	buf := make([]byte, udpscan.MaxPacketSize)

	for {
		n, hostAddress, err := listener.ReadFromUDP(buf)
		if err != nil {
			break // 超时或错误
		}

		var resp udpscan.Response
		if json.Unmarshal(buf[:n], &resp) != nil {
			// 局域网里可能有别的程序往这个端口发东西，认不出来就跳过是对的。
			// 但装满整个缓冲区还解析失败，就不是噪声而是被截断了，这个必须说一声。
			if n == len(buf) {
				fmt.Fprintf(os.Stderr, "警告: 来自 %s 的应答已达 %d 字节上限, 可能被截断而丢弃\n", hostAddress.IP, n)
			}
			continue
		}

		// 按应答来源 IP 去重：同一台机器可能从多个广播地址各收到一次探测
		ip := hostAddress.IP.String()
		if seen[ip] {
			continue
		}
		seen[ip] = true

		// 昵称过滤放在去重之后：昵称可能重复，匹配到几台就报几台，不假设唯一
		if wantedName != "" && resp.Name != wantedName {
			continue
		}

		host := udpscan.NewHost(ip, &resp)
		hosts = append(hosts, host)
		if !asJSON {
			fmt.Printf("发现: %-15s %s\n", host.IP, host.Name)
		}
	}

	result := udpscan.ScanResult{
		Broadcasts:     reached,
		Port:           targetPort,
		TimeoutSeconds: timeout,
		StartedAt:      startedAt,
		WantedName:     wantedName,
		Count:          len(hosts),
		Hosts:          hosts,
	}

	if asJSON {
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Fprintln(os.Stderr, "序列化失败:", err)
			os.Exit(1)
		}
		fmt.Println(string(data))
	} else {
		printResultText(&result)
	}

	// 一台都没扫到时以非零码退出，脚本里可以直接用 if 判断
	if result.Count == 0 {
		os.Exit(1)
	}
}

// printResultText 输出给人看的扫描结果。
func printResultText(result *udpscan.ScanResult) {
	fmt.Println("--------------------")
	if result.WantedName != "" {
		fmt.Printf("匹配 %q 的主机: %d 台\n", result.WantedName, result.Count)
	} else {
		fmt.Printf("共发现 %d 台主机\n", result.Count)
	}

	// 把拼好的 ssh 命令单独列一遍，可以直接复制执行
	var sshCommands []string
	for _, item := range result.Hosts {
		if item.SSHCommand != "" {
			sshCommands = append(sshCommands, item.SSHCommand)
		}
	}
	if len(sshCommands) > 0 {
		fmt.Println()
		for _, line := range sshCommands {
			fmt.Println(line)
		}
	}
}
