[![GitHub Workflow Status (branch)](https://img.shields.io/github/actions/workflow/status/yylego/udpscan/release.yml?branch=main&label=BUILD)](https://github.com/yylego/udpscan/actions/workflows/release.yml?query=branch%3Amain)
[![GoDoc](https://pkg.go.dev/badge/github.com/yylego/udpscan)](https://pkg.go.dev/github.com/yylego/udpscan)
[![Coverage Status](https://img.shields.io/coveralls/github/yylego/udpscan/main.svg)](https://coveralls.io/github/yylego/udpscan?branch=main)
[![Supported Go Versions](https://img.shields.io/badge/Go-1.25+-lightgrey.svg)](https://go.dev/)
[![GitHub Release](https://img.shields.io/github/release/yylego/udpscan.svg)](https://github.com/yylego/udpscan/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/yylego/udpscan)](https://goreportcard.com/report/github.com/yylego/udpscan)

# udpscan

Detect LAN hosts via UDP broadcast.

---

<!-- TEMPLATE (EN) BEGIN: LANGUAGE NAVIGATION -->

## CHINESE README

[中文说明](README.zh.md)

<!-- TEMPLATE (EN) END: LANGUAGE NAVIGATION -->

## Installation

```bash
go install github.com/yylego/udpscan/cmd/udpscan@latest
```

## Usage

### Start Host Detection Service

Run on hosts you want to detect:

```bash
udpscan server --name xiaozhixiang
```

Options:
- `--name, -n` : Host nickname (required)
- `--port, -p` : Listen port (default: 42388)

### Scan LAN Hosts

Scan and detect hosts in the same subnet:

```bash
udpscan client --broadcast 192.168.1.255
```

Options:
- `--broadcast, -b` : Broadcast address (default: 255.255.255.255)
- `--port, -p` : Target port (default: 42388)
- `--timeout, -t` : Scan timeout in seconds (default: 3)

## Example

```bash
# On Ubuntu (start detection service)
$ udpscan server --name xiaozhixiang
服务端启动 [xiaozhixiang] 监听端口 42388

# On Mac (scan subnet)
$ udpscan client --broadcast 172.16.91.255
扫描中... (广播: 172.16.91.255, 端口: 42388, 超时: 3s)
--------------------
发现: 172.16.91.98 -> xiaozhixiang
--------------------
共发现 1 台主机
```

## How It Works

1. **Host Detection Service** listens on a UDP port and responds with its nickname when it receives a scan request
2. **Scan Command** sends a UDP broadcast to the subnet and collects responses from active hosts
3. Results show IP addresses and corresponding nicknames of detected hosts

---

<!-- TEMPLATE (EN) BEGIN: STANDARD PROJECT FOOTER -->
<!-- VERSION 2025-11-25 03:52:28.131064 +0000 UTC -->

## 📄 License

MIT License - see [LICENSE](LICENSE).

---

## 💬 Contact & Feedback

Contributions are welcome! Report bugs, suggest features, and contribute code:

- 🐛 **Mistake reports?** Open an issue on GitHub with reproduction steps
- 💡 **Fresh ideas?** Create an issue to discuss
- 📖 **Documentation confusing?** Report it so we can improve
- 🚀 **Need new features?** Share the use cases to help us understand requirements
- ⚡ **Performance issue?** Help us optimize through reporting slow operations
- 🔧 **Configuration problem?** Ask questions about complex setups
- 📢 **Follow project progress?** Watch the repo to get new releases and features
- 🌟 **Success stories?** Share how this package improved the workflow
- 💬 **Feedback?** We welcome suggestions and comments

---

## 🔧 Development

New code contributions, follow this process:

1. **Fork**: Fork the repo on GitHub (using the webpage UI).
2. **Clone**: Clone the forked project (`git clone https://github.com/yourname/repo-name.git`).
3. **Navigate**: Navigate to the cloned project (`cd repo-name`)
4. **Branch**: Create a feature branch (`git checkout -b feature/xxx`).
5. **Code**: Implement the changes with comprehensive tests
6. **Testing**: (Golang project) Ensure tests pass (`go test ./...`) and follow Go code style conventions
7. **Documentation**: Update documentation to support client-facing changes
8. **Stage**: Stage changes (`git add .`)
9. **Commit**: Commit changes (`git commit -m "Add feature xxx"`) ensuring backward compatible code
10. **Push**: Push to the branch (`git push origin feature/xxx`).
11. **PR**: Open a merge request on GitHub (on the GitHub webpage) with detailed description.

Please ensure tests pass and include relevant documentation updates.

---

## 🌟 Support

Welcome to contribute to this project via submitting merge requests and reporting issues.

**Project Support:**

- ⭐ **Give GitHub stars** if this project helps you
- 🤝 **Share with teammates** and (golang) programming friends
- 📝 **Write tech blogs** about development tools and workflows - we provide content writing support
- 🌟 **Join the ecosystem** - committed to supporting open source and the (golang) development scene

**Have Fun Coding with this package!** 🎉🎉🎉

<!-- TEMPLATE (EN) END: STANDARD PROJECT FOOTER -->

---

## GitHub Stars

[![Stargazers](https://starchart.cc/yylego/udpscan.svg?variant=adaptive)](https://starchart.cc/yylego/udpscan)
