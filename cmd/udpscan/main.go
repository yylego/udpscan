package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/yylego/must"
	"github.com/yylego/udpscan"
)

var (
	name      string
	port      int
	broadcast string
	timeout   int
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "udpscan",
		Short: "UDP 局域网主机发现工具",
	}

	serverCmd := &cobra.Command{
		Use:   "server",
		Short: "启动服务端，监听并回复自己的昵称",
		Run:   runServer,
	}
	serverCmd.Flags().StringVarP(&name, "name", "n", "", "主机昵称 (必填)")
	serverCmd.Flags().IntVarP(&port, "port", "p", udpscan.ScanPort, "监听端口")
	must.Done(serverCmd.MarkFlagRequired("name"))

	clientCmd := &cobra.Command{
		Use:   "client",
		Short: "扫描局域网内的主机",
		Run:   runClient,
	}
	clientCmd.Flags().StringVarP(&broadcast, "broadcast", "b", "255.255.255.255", "广播地址")
	clientCmd.Flags().IntVarP(&port, "port", "p", udpscan.ScanPort, "扫描端口")
	clientCmd.Flags().IntVarP(&timeout, "timeout", "t", 3, "扫描超时(秒)")

	rootCmd.AddCommand(serverCmd, clientCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runServer(cmd *cobra.Command, args []string) {
	addr := &net.UDPAddr{IP: net.IPv4(0, 0, 0, 0), Port: port}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("启动失败:", err)
		os.Exit(1)
	}
	defer func() { _ = conn.Close() }()

	fmt.Printf("服务端启动 [%s] 监听端口 %d\n", name, port)

	buf := make([]byte, 1024)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("读取错误:", err)
			continue
		}

		msg := string(buf[:n])
		if msg == udpscan.ScanSign {
			resp := udpscan.Response{Name: name, Time: time.Now()}
			data, _ := json.Marshal(resp)
			_, _ = conn.WriteToUDP(data, remoteAddr)
			fmt.Printf("回复 %s\n", remoteAddr.IP)
		}
	}
}

func runClient(cmd *cobra.Command, args []string) {
	broadcastIP := net.ParseIP(broadcast)
	if broadcastIP == nil {
		fmt.Println("无效的广播地址:", broadcast)
		os.Exit(1)
	}

	// 先监听随机端口，用于接收回复
	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(0, 0, 0, 0), Port: 0})
	if err != nil {
		fmt.Println("监听失败:", err)
		os.Exit(1)
	}
	defer func() { _ = listener.Close() }()

	// 发送广播到服务端端口
	serverAddr := &net.UDPAddr{IP: broadcastIP, Port: port}
	_, err = listener.WriteToUDP([]byte(udpscan.ScanSign), serverAddr)
	if err != nil {
		fmt.Println("发送失败:", err)
		os.Exit(1)
	}

	must.Done(listener.SetReadDeadline(time.Now().Add(time.Duration(timeout) * time.Second)))

	fmt.Printf("扫描中... (广播: %s, 端口: %d, 超时: %ds)\n", broadcast, port, timeout)
	fmt.Println("--------------------")

	hosts := make(map[string]string)
	buf := make([]byte, 1024)

	for {
		n, remoteAddr, err := listener.ReadFromUDP(buf)
		if err != nil {
			break // 超时或错误
		}

		var resp udpscan.Response
		if json.Unmarshal(buf[:n], &resp) == nil {
			ip := remoteAddr.IP.String()
			hosts[ip] = resp.Name
			fmt.Printf("发现: %s -> %s\n", ip, resp.Name)
		}
	}

	fmt.Println("--------------------")
	fmt.Printf("共发现 %d 台主机\n", len(hosts))
}
