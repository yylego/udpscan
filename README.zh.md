[![GitHub Workflow Status (branch)](https://img.shields.io/github/actions/workflow/status/yylego/udpscan/release.yml?branch=main&label=BUILD)](https://github.com/yylego/udpscan/actions/workflows/release.yml?query=branch%3Amain)
[![GoDoc](https://pkg.go.dev/badge/github.com/yylego/udpscan)](https://pkg.go.dev/github.com/yylego/udpscan)
[![Coverage Status](https://img.shields.io/coveralls/github/yylego/udpscan/main.svg)](https://coveralls.io/github/yylego/udpscan?branch=main)
[![Supported Go Versions](https://img.shields.io/badge/Go-1.25+-lightgrey.svg)](https://go.dev/)
[![GitHub Release](https://img.shields.io/github/release/yylego/udpscan.svg)](https://github.com/yylego/udpscan/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/yylego/udpscan)](https://goreportcard.com/report/github.com/yylego/udpscan)

# udpscan

通过 UDP 广播检测局域网主机。

---

<!-- TEMPLATE (ZH) BEGIN: LANGUAGE NAVIGATION -->

## 英文文档

[ENGLISH README](README.md)

<!-- TEMPLATE (ZH) CLOSE: LANGUAGE NAVIGATION -->

## 核心特性

- 🎯 **用名字代替地址**：主机自报昵称，找机器靠名字，不必记那串谁也背不下来的地址
- ⚡ **直接可用的 ssh 命令**：应答带上登录用户名与 sshd 端口，扫描端用"应答实际来自哪个地址"拼出命令
- 🌐 **覆盖每一张网卡**：逐个网卡按各自的广播地址探测，点对点隧道按能力标志跳过而不是按名字排除
- 📦 **为脚本而生**：`--json` 输出结构化结果，退出码表明有没有扫到，进度信息不会污染标准输出
- 🔒 **对防火墙友好**：固定源端口之后，放行规则可从一整段临时端口收窄到一个端口
- 🩺 **故障不会静默**：应答超长、参数非法、协议不匹配都会明确报出来，而不是被无声吞掉

---

## 安装

```bash
go install github.com/yylego/udpscan/cmd/udpscan@latest
```

## 使用方法

### 启动主机检测服务

在需要被检测的主机上运行：

```bash
udpscan server --name xiaozhixiang
```

参数说明：
- `--name, -n` ：主机昵称（必填）
- `--username, -u` ：供对方 ssh 登录的用户名（留空则取当前账号）
- `--ssh-port` ：本机 sshd 监听的端口（默认：22）
- `--port, -p` ：监听端口（默认：42388）

⚠️ 服务以 root 常驻时（LaunchDaemon、init.d），`--username` 留空会**得不到用户名**——因为自动取到的 `root` 几乎不是对方想登录的账号，这种情况下宁可不报也不报错的。**用服务方式部署时请显式指定 `--username`。**

### 扫描局域网主机

扫描本机所在的各个子网：

```bash
udpscan client
```

参数说明：
- `--broadcast, -b` ：广播地址（留空则自动枚举所有网卡）
- `--name, -n` ：只保留昵称匹配的主机（昵称可能重复，会报出匹配到的台数）
- `--json` ：输出 JSON，便于脚本处理
- `--port, -p` ：**对方**监听的端口，探测包发往这里（默认：42388）
- `--source-port` ：**本机**发包与收应答用的端口（默认 0 = 随机；固定后防火墙只需放行一个口，见「部署须知」）
- `--timeout, -t` ：扫描超时秒数（默认：3）

**退出码**：扫到至少一台返回 `0`，一台都没扫到返回 `1`，因此可以直接写进条件判断：

```bash
if udpscan client --name xiaozhixiang > /dev/null; then echo "在线"; else echo "不在线"; fi
```

### 查看版本与协议标识

```bash
udpscan version
```

```
udpscan     v0.0.1
协议魔术字  UDPSCAN
默认端口    42388
Go 版本     go1.26.5
```

⭐ 两端能不能对上话，取决于**协议魔术字**而不是版本号。两台机器各跑一次这个命令一比对，立刻能判断是不是协议不兼容——那种情况下双方都不会报任何错（详见下面「部署须知」）。

## 使用示例

```bash
# 在 Ubuntu 上（启动检测服务）
$ udpscan server --name xiaozhixiang --username yangyile
服务端启动 [xiaozhixiang] 监听端口 42388
对外通告的登录方式: yangyile@本机 (sshd 端口 22)

# 在 Mac 上（扫描，自动覆盖所有网卡）
$ udpscan client
扫描中... (广播: 172.16.91.255 10.42.0.255, 端口: 42388, 超时: 3s)
--------------------
发现: 172.16.91.98    xiaozhixiang
--------------------
共发现 1 台主机

ssh yangyile@172.16.91.98
```

末尾那行 ssh 命令可以直接复制执行——这正是扫描之后通常要做的下一件事。

给脚本用时加 `--json`：

```bash
$ udpscan client --json
{
  "broadcasts": ["172.16.91.255", "10.42.0.255"],
  "port": 42388,
  "timeout_seconds": 3,
  "started_at": "2026-08-08T02:42:18+08:00",
  "count": 1,
  "hosts": [
    {
      "ip": "172.16.91.98",
      "name": "xiaozhixiang",
      "username": "yangyile",
      "ssh_port": 22,
      "ssh_command": "ssh yangyile@172.16.91.98",
      "time": "2026-08-08T02:42:18+08:00"
    }
  ]
}
```

⭐ 开头那几个概览字段在**没扫到任何主机时最有用**：`broadcasts` 告诉你探测包究竟发往了哪些地址，据此能分辨是"确实没有主机应答"还是"包压根发错了网卡"。

## 工作原理

1. **主机检测服务** 监听 UDP 端口，收到扫描请求时回复自身昵称、登录用户名和 sshd 端口
2. **扫描命令** 枚举本机每张网卡各自的广播地址，逐个发送 UDP 广播，收集活跃主机的响应
3. 结果展示检测到的主机 IP 地址、昵称，以及拼好的 ssh 命令

⭐ **ssh 命令里的 IP 由扫描端决定，不是被扫端自报的。** 一台多网卡的主机并不知道自己是通过哪个地址被看到的，只有扫描端手里那个"实际收到应答的来源地址"才保证可达。

## 部署须知

以下三点都会让扫描"静默失败"——不报错、不超时报警，只是扫不到主机，因此很难往这些方向想。部署前先看完，比事后抓包省事得多。

### 防火墙必须放行两个方向

只放行"目的端口 = 扫描端口"是不够的。那只放行了**别人扫我**的请求包，没放行**我扫别人**的应答包。

- **症状**：别人能扫到这台机器，这台机器却扫不到任何人
- **原因**：广播请求发往 `x.x.x.255`，而应答来自具体主机 IP（如 `x.x.x.190`），源地址对不上号，连接跟踪判不出这是已建立连接的回程，于是被默认策略丢弃。同时应答包的目的端口是随机临时端口、不是扫描端口，所以"目的端口"那条规则也匹配不上
- **对策**：按**源端口 = 扫描端口**再放行一条。可把目的端口限定在本机临时端口范围内（Linux 见 `net.ipv4.ip_local_port_range`），只让扫描进程的随机端口收得到，不影响本机其它固定端口

#### ⭐ 用 `--source-port` 把放行范围收窄到一个端口

上面那条规则的目的端口是**一整段临时端口范围**：

```bash
# 放行了 32768-60999，足足 28232 个端口
-A INPUT -p udp --sport 42388 --dport 32768:60999 -j ACCEPT
```

范围之所以要开这么大，是因为客户端默认用**随机端口**收应答，事先不知道会用哪个。而这条规则的判据是"源端口 = 42388"——**源端口是发送方随便填的**，于是任何人都能借这个口子够到本机那两万八千个端口上的其它服务。

给客户端固定一个收包端口，规则就能精确到一个口：

```bash
# 扫描时固定用 42389 收应答
udpscan client --source-port 42389

# 防火墙只需放行这一个端口
-A INPUT -p udp --sport 42388 --dport 42389 -j ACCEPT
```

**28232 个端口 → 1 个端口。**

⚠️ 代价是同一台机器上**不能同时跑两个扫描**（端口会被占用）。此时命令会**直接报错退出、并提示去掉该参数**——**故意不自动退回随机端口**：那样防火墙规则会突然匹配不上，变成"有时扫得到、有时扫不到"的诡异现象，比直接失败难查得多。

### 别用受限广播地址，交给自动枚举

`255.255.255.255` 是受限广播地址，内核按路由表决定从哪个接口发出。机器上若存在 VPN、代理或虚拟网卡，它很可能从那些接口出去，根本到不了目标网段——**而且发送不会报错**。

**现在默认就绕开了这个坑**：不带 `-b` 时，客户端会枚举本机每张网卡各自的定向广播地址（如 `192.168.1.255`）并逐个发送，每个所在网段都覆盖得到。VPN、代理建的点对点隧道（`tun` / `utun`）不具备广播能力，会被自动跳过——判断的是接口的广播能力标志，不是按名字硬编码排除，所以隧道叫什么名字都不影响。

`-b` 现在只在两种情况下才需要：想**只扫某一个网段**，或者要发往**本机网卡之外**的广播地址。⛔ 不建议手动传 `255.255.255.255`，那等于把上面这个坑又请回来了。

### 两端版本必须一致

握手依靠一个约定的魔术字符串，两端严格相等才会应答。**魔术字一旦变更，新旧版本之间完全静默**：请求照发、服务端照收，只是不匹配就直接丢弃，双方都不会打印任何异常。

升级时把所有机器上的二进制一起换掉。排查这类问题时，直接抓一次包看请求内容比读日志更快。

---

<!-- TEMPLATE (ZH) BEGIN: STANDARD PROJECT FOOTER -->
<!-- VERSION 2025-11-25 03:52:28.131064 +0000 UTC -->

## 📄 许可证类型

MIT 许可证 - 详见 [LICENSE](LICENSE)。

---

## 💬 联系与反馈

非常欢迎贡献代码！报告 BUG、建议功能、贡献代码：

- 🐛 **问题报告？** 在 GitHub 上提交问题并附上重现步骤
- 💡 **新颖思路？** 创建 issue 讨论
- 📖 **文档疑惑？** 报告问题，帮助我们完善文档
- 🚀 **需要功能？** 分享使用场景，帮助理解需求
- ⚡ **性能瓶颈？** 报告慢操作，协助解决性能问题
- 🔧 **配置困扰？** 询问复杂设置的相关问题
- 📢 **关注进展？** 关注仓库以获取新版本和功能
- 🌟 **成功案例？** 分享这个包如何改善工作流程
- 💬 **反馈意见？** 欢迎提出建议和意见

---

## 🔧 代码贡献

新代码贡献，请遵循此流程：

1. **Fork**：在 GitHub 上 Fork 仓库（使用网页界面）
2. **克隆**：克隆 Fork 的项目（`git clone https://github.com/yourname/repo-name.git`）
3. **导航**：进入克隆的项目（`cd repo-name`）
4. **分支**：创建功能分支（`git checkout -b feature/xxx`）
5. **编码**：实现您的更改并编写全面的测试
6. **测试**：（Golang 项目）确保测试通过（`go test ./...`）并遵循 Go 代码风格约定
7. **文档**：面向用户的更改需要更新文档
8. **暂存**：暂存更改（`git add .`）
9. **提交**：提交更改（`git commit -m "Add feature xxx"`）确保向后兼容的代码
10. **推送**：推送到分支（`git push origin feature/xxx`）
11. **PR**：在 GitHub 上打开 Merge Request（在 GitHub 网页上）并提供详细描述

请确保测试通过并包含相关的文档更新。

---

## 🌟 项目支持

非常欢迎通过提交 Merge Request 和报告问题来贡献此项目。

**项目支持：**

- ⭐ **给予星标**如果项目对您有帮助
- 🤝 **分享项目**给团队成员和（golang）编程朋友
- 📝 **撰写博客**关于开发工具和工作流程 - 我们提供写作支持
- 🌟 **加入生态** - 致力于支持开源和（golang）开发场景

**祝你用这个包编程愉快！** 🎉🎉🎉

<!-- TEMPLATE (ZH) CLOSE: STANDARD PROJECT FOOTER -->

---

## GitHub 标星点赞

[![标星点赞](https://starchart.cc/yylego/udpscan.svg?variant=adaptive)](https://starchart.cc/yylego/udpscan)
