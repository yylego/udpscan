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

<!-- TEMPLATE (ZH) END: LANGUAGE NAVIGATION -->

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
- `--port, -p` ：监听端口（默认：42388）

### 扫描局域网主机

扫描同一子网内的主机：

```bash
udpscan client --broadcast 192.168.1.255
```

参数说明：
- `--broadcast, -b` ：广播地址（默认：255.255.255.255）
- `--port, -p` ：目标端口（默认：42388）
- `--timeout, -t` ：扫描超时秒数（默认：3）

## 使用示例

```bash
# 在 Ubuntu 上（启动检测服务）
$ udpscan server --name xiaozhixiang
服务端启动 [xiaozhixiang] 监听端口 42388

# 在 Mac 上（扫描子网）
$ udpscan client --broadcast 172.16.91.255
扫描中... (广播: 172.16.91.255, 端口: 42388, 超时: 3s)
--------------------
发现: 172.16.91.98 -> xiaozhixiang
--------------------
共发现 1 台主机
```

## 工作原理

1. **主机检测服务** 监听 UDP 端口，收到扫描请求时回复自身昵称
2. **扫描命令** 向子网发送 UDP 广播，收集活跃主机的响应
3. 结果展示检测到的主机 IP 地址和对应昵称

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

<!-- TEMPLATE (ZH) END: STANDARD PROJECT FOOTER -->

---

## GitHub 标星点赞

[![标星点赞](https://starchart.cc/yylego/udpscan.svg?variant=adaptive)](https://starchart.cc/yylego/udpscan)
